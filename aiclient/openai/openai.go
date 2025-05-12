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

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/breaker"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/retrier"
	"github.com/alnah/fla/tripper"
)

const (
	apiKeyFromEnv = ai.EnvAPIKeyOpenAI
	gateway       = "https://api.openai.com/v1/"
	pathChat      = "chat/completions"
	pathTTS       = "audio/speech"
	pathSTT       = "audio/transcriptions"
)

type model string

func (m model) String() string { return string(m) }

const (
	// ModelReasoning selects the compact reasoning-optimized LLM.
	ModelReasoning model = "o4-mini"
	// ModelFlagship selects the flagship LLM for highest capability.
	ModelFlagship model = "gpt-4.1"
	// ModelCostOptimized selects a lower-cost variant of the flagship LLM.
	ModelCostOptimized model = "gpt-4.1-nano"
	// ModelTextToSpeech selects the LLM optimized for generating speech output.
	ModelTextToSpeech model = "gpt-4o-mini-tts"
	// ModelSpeechToText selects the LLM optimized for transcribing audio.
	ModelSpeechToText model = "gpt-4o-transcribe"
)

// Base centralizes common configuration for OpenAI API calls.
type Base struct {
	Model         model `json:"model"`
	hc            *http.Client
	log           *logger.Logger
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
func WithLogger[T hasBase](log *logger.Logger) option[T] {
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
	Messages    []ai.Message `json:"messages"`
	MaxTokens   int          `json:"max_completion_tokens,omitempty"`
	Temperature float32      `json:"temperature,omitempty"`
	*Base
	systemOnce sync.Once
}

// GetBase satisfies hasBase interface to use shared options.
func (c *Chat) GetBase() *Base { return c.Base }

// WithMessages initializes the chat history to send in the request.
func WithMessages(m []ai.Message) option[*Chat] { return func(c *Chat) { c.Messages = m } }

// WithMaxTokens limits how many tokens the model may generate.
func WithMaxTokens(n int) option[*Chat] {
	return func(c *Chat) { c.MaxTokens = n }
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

// SetSystem guides the model with general instructions.
func (c *Chat) SetSystem(instructions string) *Chat {
	c.systemOnce.Do(func() {
		c.Messages = append(c.Messages, ai.Message{
			Role:    ai.RoleSystem,
			Content: instructions,
		})
	})
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

// Completion sends the assembled chat request to the OpenAI API.
func (c *Chat) Completion() (ai.Completion, error) {
	if len(c.Messages) == 0 {
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "messages required", nil)
	}

	byt, err := json.Marshal(c)
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(c.ctx, c.method, c.url, bytes.NewBuffer(byt))
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "failed to build http request", err)
	}

	c.transportOnce.Do(func() {
		orig := c.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		c.hc.Transport = tripper.Chain(
			orig,
			tripper.AddHeader("Content-Type", "application/json"),
			tripper.AddHeader("Authorization", "Bearer "+c.Base.apiKey),
			tripper.AddHeader("User-Agent", "Fla/1.0"),
			tripper.UseStatusClassifier(ai.ProviderOpenAI),
			tripper.UseCircuitBreaker(breaker.New()),
			tripper.UseRetrier(retrier.New(), isRetryable),
			tripper.UseLogger(c.Base.log),
		)
	})

	res, err := c.hc.Do(req)
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "failed to read response body", err)
	}

	if res.StatusCode != http.StatusOK {
		httpErr := ai.NewHTTPError(ai.ProviderOpenAI, res)
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "http request failed", httpErr)
	}

	var completion ai.Completion
	if err = json.Unmarshal(byt, &completion); err != nil {
		return ai.Completion{}, ai.NewChatError(ai.ProviderOpenAI, "failed to unmarshal response body", err)
	}

	return completion, nil
}

type voice string

func (v voice) String() string { return string(v) }

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

// Audio sends the TTS request and returns the synthesized audio bytes.
func (t *TTS) Audio() ([]byte, error) {
	if t.Input == "" {
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "text input required", nil)
	}

	byt, err := json.Marshal(t)
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(t.ctx, t.method, t.url, bytes.NewBuffer(byt))
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "failed to build http request", err)
	}

	t.transportOnce.Do(func() {
		orig := t.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		t.hc.Transport = tripper.Chain(
			orig,
			tripper.AddHeader("Content-Type", "application/json"),
			tripper.AddHeader("Authorization", "Bearer "+t.Base.apiKey),
			tripper.AddHeader("User-Agent", "Fla/1.0"),
			tripper.UseStatusClassifier(ai.ProviderOpenAI),
			tripper.UseCircuitBreaker(breaker.New()),
			tripper.UseRetrier(retrier.New(), isRetryable),
			tripper.UseLogger(t.log),
		)
	})

	res, err := t.hc.Do(req)
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "failed to read response body", err)
	}

	if res.StatusCode != http.StatusOK {
		httpErr := ai.NewHTTPError(ai.ProviderOpenAI, res)
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "http request failed", httpErr)
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

type fileFmt string

func (ff fileFmt) Stirng() string { return string(ff) }

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
func (s *STT) Transcript() (ai.Transcription, error) {
	if s.File == nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "file required", nil)
	}
	if s.FilePath == "" {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "filepath required", nil)
	}
	if s.Language == "" {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "source language required", nil)
	}

	bodyBuf := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(bodyBuf)
	defer func() { _ = multipartWriter.Close() }()

	part, err := multipartWriter.CreateFormFile("file", s.FilePath)
	if err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, fmt.Sprintf("failed to create form file for path: %q", s.FilePath), err)
	}
	if _, err := s.File.Seek(0, io.SeekStart); err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, fmt.Sprintf("failed to seek file for path %q to start", s.FilePath), err)
	}
	if _, err := io.Copy(part, s.File); err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, fmt.Sprintf("failed to copy file contents to multipart writer for path %q", s.FilePath), err)
	}
	if err := multipartWriter.WriteField("model", s.Base.Model.String()); err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to write model field to multipart body", err)
	}

	if err := multipartWriter.WriteField("language", s.Language); err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to write language field to multipart body", err)
	}
	if err := multipartWriter.Close(); err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to finalize multipart body writer", err)
	}

	req, err := http.NewRequestWithContext(s.Base.ctx, s.method, s.url, bodyBuf)
	if err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to build http request", err)
	}

	s.Base.transportOnce.Do(func() {
		orig := s.Base.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		s.Base.hc.Transport = tripper.Chain(
			orig,
			tripper.AddHeader("Content-Type", multipartWriter.FormDataContentType()),
			tripper.AddHeader("Authorization", "Bearer "+s.Base.apiKey),
			tripper.AddHeader("User-Agent", "Fla/1.0"),
			tripper.UseStatusClassifier(ai.ProviderOpenAI),
			tripper.UseCircuitBreaker(breaker.New()),
			tripper.UseRetrier(retrier.New(), isRetryable),
			tripper.UseLogger(s.Base.log),
		)
	})

	res, err := s.Base.hc.Do(req)
	if err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err := io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to read response body", err)
	}

	if res.StatusCode != http.StatusOK {
		httpErr := ai.NewHTTPError(ai.ProviderOpenAI, res)
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "http request failed", httpErr)
	}

	var body ai.Transcription
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return ai.Transcription{}, ai.NewSTTError(ai.ProviderOpenAI, "failed to decode response body", err)
	}

	return body, nil
}

const (
	errRateLimited ai.ErrType = "rate_limited"
	errServer      ai.ErrType = "server_error"
	errTimeout     ai.ErrType = "timeout_error"
	errUnavailable ai.ErrType = "unavailable"
)

var retryable = map[ai.ErrType]struct{}{
	errRateLimited: {},
	errServer:      {},
	errTimeout:     {},
	errUnavailable: {},
}

var isRetryable = ai.NewRetryClassifier(retryable)
