package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/breaker"
	"github.com/alnah/fla/ai/retrier"
	"github.com/alnah/fla/ai/transport"
	"github.com/alnah/fla/clog"
)

const (
	apiKeyFromEnv = "OPENAI_API_KEY" // #nosec G101: this is only the ENV var name, not a secret
	gateway       = "https://api.openai.com/v1/"
	pathChat      = "chat/completions"
	pathTTS       = "audio/speech"
	pathSTT       = "audio/transcriptions"
)

type model string

func (m model) String() string { return string(m) }

const (
	// ModelReasoning selects the compact reasoning-optimized LLM.
	ModelReasoning = "o4-mini"
	// ModelFlagship selects the flagship LLM for highest capability.
	ModelFlagship = "gpt-4.1"
	// ModelCostOptimized selects a lower-cost variant of the flagship LLM.
	ModelCostOptimized = "gpt-4.1-nano"
	// ModelTextToSpeech selects the LLM optimized for generating speech output.
	ModelTextToSpeech = "gpt-4o-mini-tts"
	// ModelSpeechToText selects the LLM optimized for transcribing audio.
	ModelSpeechToText = "gpt-4o-transcribe"
)

// Base centralizes common configuration for OpenAI API calls.
type Base struct {
	Model         model `json:"model"`
	hc            *http.Client
	log           *clog.Logger
	ctx           context.Context
	transportOnce sync.Once
	apiKey        string
	method        string
	url           string
}

type hasBase interface{ GetBase() *Base }

type option[T hasBase] func(T)

// WithModel sets which model to call.
func WithModel[T hasBase](m model) option[T] { return func(t T) { (t).GetBase().Model = m } }

// WithLogger attaches structured logging of request/response events.
func WithLogger[T hasBase](log *clog.Logger) option[T] {
	return func(t T) { (t).GetBase().log = log }
}

// WithHTTPClient overrides the default HTTP client for custom transport configurations.
func WithHTTPClient[T hasBase](hc *http.Client) option[T] {
	return func(t T) { (t).GetBase().hc = hc }
}

// WithContext supplies a context for cancellation and deadlines.
func WithContext[T hasBase](ctx context.Context) option[T] {
	return func(t T) { (t).GetBase().ctx = ctx }
}

// WithAPIKey provides the OpenAI authentication token explicitly.
func WithAPIKey[T hasBase](s string) option[T] { return func(t T) { (t).GetBase().apiKey = s } }

// WithURL directs requests to a custom endpoint, useful for testing.
func WithURL[T hasBase](s string) option[T] { return func(t T) { (t).GetBase().url = s } }

// Chat manages a conversation request to the OpenAI API.
type Chat struct {
	Messages            []ai.Message `json:"messages"`
	MaxCompletionTokens int          `json:"max_completion_tokens,omitempty"`
	Temperature         float32      `json:"temperature,omitempty"`
	*Base
}

// GetBase satisfies hasBase interface to use shared options.
func (c *Chat) GetBase() *Base { return c.Base }

// WithMessages initializes the chat history to send in the request.
func WithMessages(m []ai.Message) option[*Chat] { return func(c *Chat) { c.Messages = m } }

// WithMaxCompletionTokens limits how many tokens the model may generate.
func WithMaxCompletionTokens(n int) option[*Chat] {
	return func(c *Chat) { c.MaxCompletionTokens = n }
}

// WithTemperature adjusts response variability; higher values yield more diverse outputs.
func WithTemperature(n float32) option[*Chat] { return func(c *Chat) { c.Temperature = n } }

// NewChat constructs a chat client. Defaults model to cost optimized.
// Callers can override settings via options to customize transcription.
func NewChat(opts ...option[*Chat]) *Chat {
	c := &Chat{
		Messages: []ai.Message{},
		Base: &Base{
			Model:  ModelCostOptimized,
			hc:     &http.Client{Timeout: 30 * time.Second},
			ctx:    context.Background(),
			apiKey: os.Getenv(apiKeyFromEnv),
			method: http.MethodPost,
			url:    gateway + pathChat,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// AddMessage appends a new role/content pair to the conversation history.
// It returns the same chat to enable fluent chaining.
func (c *Chat) AddMessage(role ai.Role, content string) *Chat {
	c.Messages = append(c.Messages, ai.Message{
		Role:    role,
		Content: content,
	})
	return c
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatMessage struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Refusal *string `json:"refusal,omitempty"`
}

type completion struct {
	ID      string       `json:"id"`
	Choices []chatChoice `json:"choices"`
}

// Content extracts the main text of the first choice, or returns empty
// if no content is available.
func (res *completion) Content() string {
	if len(res.Choices) == 0 || res.Choices[0].Message.Content == "" {
		return ""
	}
	return res.Choices[0].Message.Content
}

// OpenAICompletionError indicates a request failure due to invalid input
// or an API-side error, preserving the descriptive error message.
type OpenAICompletionError struct{ message string }

func (e *OpenAICompletionError) Error() string {
	return fmt.Sprintf("OpenAICompletionError: %s", e.message)
}

// Completion sends the assembled chat request to the OpenAI API.
func (c *Chat) Completion() (completion, error) {
	if len(c.Messages) == 0 {
		return completion{}, &OpenAICompletionError{message: "messages required"}
	}

	byt, err := json.Marshal(c)
	if err != nil {
		return completion{}, &OpenAICompletionError{message: fmt.Sprintf("failed to marshal chat: %v", err)}
	}

	req, err := http.NewRequestWithContext(c.Base.ctx, c.Base.method, c.Base.url, bytes.NewBuffer(byt))
	if err != nil {
		return completion{}, &OpenAICompletionError{message: fmt.Sprintf("failed to create request: %v", err)}
	}

	c.Base.transportOnce.Do(func() {
		orig := c.Base.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		c.Base.hc.Transport = transport.Chain(
			orig,
			transport.AddHeader("Content-Type", "application/json"),
			transport.AddHeader("Authorization", "Bearer "+c.Base.apiKey),
			transport.AddHeader("User-Agent", "Fla/1.0"),
			transport.ClassifyStatus,
			transport.UseCircuitBreaker(breaker.New()),
			transport.UseRetrier(retrier.New(), isRetryable),
			transport.UseLogger(c.Base.log),
		)
	})

	res, err := c.Base.hc.Do(req)
	if err != nil {
		return completion{}, &OpenAICompletionError{message: fmt.Sprintf("failed to send HTTP request: %v", err)}
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	if err != nil {
		return completion{}, &OpenAICompletionError{message: fmt.Sprintf("failed to read response body: %v", err)}
	}

	if res.StatusCode != http.StatusOK {
		var apiErr transport.APIError
		_ = json.Unmarshal(byt, &apiErr)

		return completion{}, &transport.HTTPError{
			Status:  res.StatusCode,
			Type:    apiErr.Error.Type,
			Code:    apiErr.Error.Code,
			Message: apiErr.Error.Message,
		}
	}

	var body completion
	if err = json.Unmarshal(byt, &body); err != nil {
		return completion{}, &OpenAICompletionError{message: fmt.Sprintf("failed to unmarshal response body: %v", err)}
	}

	return body, nil
}

type voice string

const (
	// Warm, and friendly.
	VoiceMaleAlloy voice = "alloy"
	// Confident, and edgy.
	VoiceMaleAsh voice = "ash"
	// Gentle, and soothing.
	VoiceFemaleBallad voice = "ballad"
	// Energetic, and expressive.
	VoiceFemaleCoral voice = "coral"
	// Clear, and neutral.
	VoiceMaleEcho voice = "echo"
	// Whimsical, and soft.
	VoiceFemaleFable voice = "fable"
	// Deep, and serious.
	VoiceMaleOnyx voice = "onyx"
	// Calm, and wise.
	VoiceFemaleSage voice = "sage"
	// Bright, and youthful.
	VoiceFemaleShimmer voice = "shimmer"
	// Poetic, and flowing.
	VoiceMaleVerse voice = "verse"
)

// TTS holds configuration for text-to-speech synthesis.
type TTS struct {
	Input        string `json:"input"`
	Voice        voice  `json:"voice"`
	Instructions string `json:"instructions,omitempty"`
	*Base
}

// GetBase satisfies hasBase interface to use shared options.
func (t *TTS) GetBase() *Base { return t.Base }

// WithInput specifies the text to be converted into speech.
func WithInput(s string) option[*TTS] { return func(t *TTS) { t.Input = s } }

// WithVoice selects the voice style for speech output.
func WithVoice(v voice) option[*TTS] { return func(t *TTS) { t.Voice = v } }

// WithInstructions adds custom guidance for pronunciation or pacing.
func WithInstructions(i string) option[*TTS] { return func(t *TTS) { t.Instructions = i } }

// NewTTS creates a TTS client. Defaults model to Text-to-Speech and voice to Fable (female).
// Callers can override settings via options to customize speech synthesis.
func NewTTS(opts ...option[*TTS]) *TTS {
	t := &TTS{
		Voice: VoiceFemaleFable,
		Base: &Base{
			Model:  ModelTextToSpeech,
			hc:     &http.Client{Timeout: 5 * time.Minute},
			ctx:    context.Background(),
			apiKey: os.Getenv(apiKeyFromEnv),
			method: http.MethodPost,
			url:    gateway + pathTTS,
		},
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

type OpenAIAudioError struct{ message string }

func (e *OpenAIAudioError) Error() string {
	return fmt.Sprintf("OpenAIAudioError: %s", e.message)
}

// Audio sends the TTS request and returns the synthesized audio bytes.
func (t *TTS) Audio() ([]byte, error) {
	if t.Input == "" {
		return nil, &OpenAIAudioError{message: "text input required"}
	}

	byt, err := json.Marshal(t)
	if err != nil {
		return nil, &OpenAIAudioError{message: fmt.Sprintf("failed to marshal tts: %v", err)}
	}

	req, err := http.NewRequestWithContext(t.Base.ctx, t.Base.method, t.Base.url, bytes.NewBuffer(byt))
	if err != nil {
		return nil, &OpenAIAudioError{message: fmt.Sprintf("failed to create request: %v", err)}
	}

	t.Base.transportOnce.Do(func() {
		orig := t.Base.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		t.Base.hc.Transport = transport.Chain(
			orig,
			transport.AddHeader("Content-Type", "application/json"),
			transport.AddHeader("Authorization", "Bearer "+t.Base.apiKey),
			transport.AddHeader("User-Agent", "Fla/1.0"),
			transport.ClassifyStatus,
			transport.UseCircuitBreaker(breaker.New()),
			transport.UseRetrier(retrier.New(), isRetryable),
			transport.UseLogger(t.Base.log),
		)
	})

	res, err := t.Base.hc.Do(req)
	if err != nil {
		return nil, &OpenAIAudioError{message: fmt.Sprintf("failed to send HTTP request: %v", err)}
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, &OpenAIAudioError{message: fmt.Sprintf("failed to read response body: %v", err)}
	}

	if res.StatusCode != http.StatusOK {
		var apiErr transport.APIError
		_ = json.Unmarshal(byt, &apiErr)

		return nil, &transport.HTTPError{
			Status:  res.StatusCode,
			Type:    apiErr.Error.Type,
			Code:    apiErr.Error.Code,
			Message: apiErr.Error.Message,
		}
	}

	return byt, nil
}

// STT holds configuration for speech-to-text transcription.
type STT struct {
	File     *os.File
	FilePath string
	Language string
	*Base
}

// GetBase satisfies hasBase interface to use shared options.
func (s *STT) GetBase() *Base { return s.Base }

// WithFile provides the file to be converted into text.
func WithFile(f *os.File) option[*STT] { return func(s *STT) { s.File = f } }

// WithFilePath indicates the file path from the file to be conveted into text.
func WithFilePath(fp string) option[*STT] { return func(s *STT) { s.FilePath = fp } }

// WithLanguage specifies the ISO-639-1 (e.g. `fr`) format to improve accuracy and latency.
func WithLanguage(l string) option[*STT] { return func(s *STT) { s.Language = l } }

// NewSTT creates a new Speech-to-Text client. Defaults to Speech-to-Text model.
// Callers can override settings via options to customize transcription.
func NewSTT(opts ...option[*STT]) *STT {
	s := &STT{
		Base: &Base{
			Model:  ModelSpeechToText,
			hc:     &http.Client{Timeout: 5 * time.Second},
			ctx:    context.Background(),
			apiKey: os.Getenv(apiKeyFromEnv),
			method: http.MethodPost,
			url:    gateway + pathSTT,
		},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

type OpenAITranscriptError struct{ message string }

func (e *OpenAITranscriptError) Error() string {
	return fmt.Sprintf("OpenAITranscriptError: %s", e.message)
}

type transcript struct {
	Text string `json:"text"`
}

func (t transcript) Content() string { return t.Text }

type fileFmt string

const (
	FileFmtFLAC fileFmt = "flac"
	FileFmtMP3  fileFmt = "mp3"
	FileFmtMP4  fileFmt = "mp4"
	FileFmtMPEG fileFmt = "mpeg"
	FileFmtMPGA fileFmt = "mpga"
	FileFmtM4A  fileFmt = "m4a"
	FileFmtOGG  fileFmt = "ogg"
	FileFmtWAV  fileFmt = "wav"
	FileFmtWEBM fileFmt = "webm"
)

// Transcript sends the Speech-to-Text request and return the transcription.
// It assembles the request with a multipart form-data.
func (s *STT) Transcript() (transcript, error) {
	if s.File == nil {
		return transcript{}, &OpenAITranscriptError{message: "file required"}
	}
	if s.FilePath == "" {
		return transcript{}, &OpenAITranscriptError{message: "filepath required"}
	}
	if s.Language == "" {
		return transcript{}, &OpenAITranscriptError{message: "source language required"}
	}

	bodyBuf := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(bodyBuf)
	defer func() { _ = multipartWriter.Close() }()

	part, err := multipartWriter.CreateFormFile("file", s.FilePath)
	if err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to create form file for path %q: %v", s.FilePath, err),
		}
	}
	if _, err := s.File.Seek(0, io.SeekStart); err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to seek file to file for path %q to start: %v", s.FilePath, err),
		}
	}
	if _, err := io.Copy(part, s.File); err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to copy file contents to multipart writer for path %q: %v", s.FilePath, err),
		}
	}
	if err := multipartWriter.WriteField("model", s.Base.Model.String()); err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to write 'model' field to multipart body: %v", err),
		}
	}

	if err := multipartWriter.WriteField("language", s.Language); err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to write 'language' field to multipart body: %v", err),
		}
	}

	if err := multipartWriter.Close(); err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to finalize multipart body writer: %v", err),
		}
	}

	req, err := http.NewRequestWithContext(s.Base.ctx, s.method, s.url, bodyBuf)
	if err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to create HTTP request: %v", err),
		}
	}

	s.Base.transportOnce.Do(func() {
		orig := s.Base.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		s.Base.hc.Transport = transport.Chain(
			orig,
			transport.AddHeader("Content-Type", multipartWriter.FormDataContentType()),
			transport.AddHeader("Authorization", "Bearer "+s.Base.apiKey),
			transport.AddHeader("User-Agent", "Fla/1.0"),
			transport.ClassifyStatus,
			transport.UseCircuitBreaker(breaker.New()),
			transport.UseRetrier(retrier.New(), isRetryable),
			transport.UseLogger(s.Base.log),
		)
	})

	res, err := s.Base.hc.Do(req)
	if err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to send HTTP request: %v", err),
		}
	}
	defer func() { _ = res.Body.Close() }()

	var body transcript
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return transcript{}, &OpenAITranscriptError{
			message: fmt.Sprintf("failed to decode response body: %v", err),
		}
	}

	return body, nil
}

const (
	errRateLimited transport.ErrType = "rate_limited"
	errServer      transport.ErrType = "server_error"
	errTimeout     transport.ErrType = "timeout_error"
	errUnavailable transport.ErrType = "unavailable"
)

var retryable = map[transport.ErrType]struct{}{
	errRateLimited: {},
	errServer:      {},
	errTimeout:     {},
	errUnavailable: {},
}

var isRetryable = transport.NewRetryClassifier(retryable)
