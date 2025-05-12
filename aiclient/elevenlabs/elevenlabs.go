package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	ai "github.com/alnah/fla/aiclient"
	"github.com/alnah/fla/aiclient/breaker"
	"github.com/alnah/fla/aiclient/retrier"
	"github.com/alnah/fla/aiclient/transport"
	"github.com/alnah/fla/logger"
)

const (
	apiKeyFromEnv = ai.EnvAPIKeyElevenLabs
	gateway       = "https://api.elevenlabs.io/v1/"
	pathTTS       = "text-to-speech/"
)

type model string

func (m model) String() string { return string(m) }

// ModelTextToSpeech selects the LLM optimized for generating speech output.
const ModelTextToSpeech model = "eleven_multilingual_v2"

type voiceID string

func (v voiceID) String() string { return string(v) }

const (
	// Middle age voice, for narration.
	VoiceMaleNicolas voiceID = "aQROLel5sQbj1vuIVi6B"
	// Young age voice, for narration and voiceover.
	VoiceMaleGuillaume voiceID = "ohItIVrXTBI80RrUECOD"
	// Middle age voice, for commercials.
	VoiceFemaleAudrey voiceID = "McVZB9hVxVSk3Equu8EH"
)

type option func(*TTS)

// WithModel sets which model to call.
func WithModel(m model) option { return func(t *TTS) { (t).Model = m } }

// WithLogger attaches structured logging of request/response events.
func WithLogger(log *logger.Logger) option { return func(t *TTS) { (t).log = log } }

// WithHTTPClient overrides the default HTTP client for custom transport configurations.
func WithHTTPClient(hc *http.Client) option { return func(t *TTS) { (t).hc = hc } }

// WithContext supplies a context for cancellation and deadlines.
func WithContext(ctx context.Context) option { return func(t *TTS) { (t).ctx = ctx } }

// WithAPIKey provides the OpenAI authentication token explicitly.
func WithAPIKey(s string) option { return func(t *TTS) { (t).apiKey = s } }

// WithURL directs requests to a custom endpoint, useful for testing.
func WithURL(s string) option { return func(t *TTS) { (t).url = s } }

// WithInput specifies the text to be converted into speech.
func WithInput(s string) option { return func(t *TTS) { t.Input = s } }

// WithVoice sets which voice to use.
func WithVoice(v voiceID) option { return func(t *TTS) { t.voice = v } }

// WithSpeed controls the speed of the generated speech. Defaults to 1.0.
func WithSpeed(n float32) option { return func(t *TTS) { t.Speed = n } }

type TTS struct {
	Input         string  `json:"text"`
	Speed         float32 `json:"speed,omitempty"` // min: 0.7, max: 1.2
	Model         model   `json:"model_id"`
	voice         voiceID
	hc            *http.Client
	log           *logger.Logger
	ctx           context.Context
	transportOnce sync.Once
	apiKey        string
	method        string
	url           string
}

// NewTTS creates a TTS client. Defaults model to Text-to-Speech and voice to Fable (female).
// Callers can override settings via options to customize speech synthesis.
func NewTTS(opts ...option) *TTS {
	t := &TTS{
		voice:  VoiceMaleNicolas,
		Model:  ModelTextToSpeech,
		hc:     &http.Client{Timeout: 5 * time.Minute},
		ctx:    context.Background(),
		apiKey: os.Getenv(apiKeyFromEnv),
		method: http.MethodPost,
		url:    gateway + pathTTS,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Audio sends the TTS request and returns the synthesized audio bytes.
func (t *TTS) Audio() ([]byte, error) {
	if t.Input == "" {
		return nil, ai.NewTTSError(ai.ProviderElevenLabs, "text input required", nil)
	}

	byt, err := json.Marshal(t)
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderElevenLabs, "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(t.ctx, t.method, t.url+t.voice.String(), bytes.NewBuffer(byt))
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderElevenLabs, "failed to build http request", err)
	}

	t.transportOnce.Do(func() {
		orig := t.hc.Transport
		if orig == nil {
			orig = http.DefaultTransport
		}
		t.hc.Transport = transport.Chain(
			orig,
			transport.AddHeader("Content-Type", "application/json"),
			transport.AddHeader("xi-api-key", t.apiKey),
			transport.AddHeader("User-Agent", "Fla/1.0"),
			transport.ClassifyStatus(ai.ProviderElevenLabs),
			transport.UseCircuitBreaker(breaker.New()),
			transport.UseRetrier(retrier.New(), isRetryable),
			transport.UseLogger(t.log),
		)
	})

	res, err := t.hc.Do(req)
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderElevenLabs, "failed to send http request", err)
	}
	defer func() { _ = res.Body.Close() }()

	byt, err = io.ReadAll(res.Body)
	res.Body = io.NopCloser(bytes.NewReader(byt))
	if err != nil {
		return nil, ai.NewTTSError(ai.ProviderElevenLabs, "failed to read response body", err)
	}

	if res.StatusCode != http.StatusOK {
		httpErr := ai.NewHTTPError(ai.ProviderOpenAI, res)
		return nil, ai.NewTTSError(ai.ProviderOpenAI, "http request failed", httpErr)
	}

	return byt, nil
}

const (
	errTooManyConcurrentRequests ai.ErrType = "too_many_concurrent_requests"
	errSystemBusy                ai.ErrType = "system_busy"
)

var retryable = map[ai.ErrType]struct{}{
	errTooManyConcurrentRequests: {},
	errSystemBusy:                {},
}

var isRetryable = ai.NewRetryClassifier(retryable)
