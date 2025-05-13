package aiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const (
	ProviderOpenAI     Provider = "openai"
	ProviderAnthropic  Provider = "anthropic"
	ProviderElevenLabs Provider = "elevenlabs"
)

type Provider string

func (p Provider) String() string { return string(p) }
func (p Provider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderElevenLabs:
		return true
	default:
		return false
	}
}
func (p Provider) Validate() error {
	if p.String() == "" {
		return fmt.Errorf("invalid provider: empty")
	}
	if !p.IsValid() {
		return fmt.Errorf("invalid provider: %s", p)
	}
	return nil
}

const (
	URLChatCompletionOpenAI    URL = "https://api.openai.com/v1/chat/completions"
	URLChatCompletionAnthropic URL = "https://api.anthropic.com/v1/messages"
	URLSpeechAudioOpenAI       URL = "https://api.openai.com/v1/audio/speech"
	URLSpeechAudioElevenLabs   URL = "https://api.elevenlabs.io/v1/text-to-speech"
	URLTranscriptionOpenAI     URL = "https://api.openai.com/v1/audio/transcriptions"
)

type URL string

func (u URL) String() string { return string(u) }
func (u URL) IsValid() bool {
	switch u {
	case URLChatCompletionOpenAI, URLChatCompletionAnthropic,
		URLSpeechAudioOpenAI, URLSpeechAudioElevenLabs,
		URLTranscriptionOpenAI:
		return true
	default:
		return false
	}
}
func (u URL) Validate() error {
	if !u.IsValid() {
		return fmt.Errorf("invalid url: %s", u)
	}
	return nil
}

const (
	EnvOpenAIAPIKey     APIKey = "OPENAI_API_KEY"     // #nosec G101
	EnvAnthropicAPIKey  APIKey = "ANTHROPIC_API_KEY"  // #nosec G101
	EnvElevenLabsAPIKey APIKey = "ELEVENLABS_API_KEY" // #nosec G101
)

type APIKey string

func (e APIKey) String() string { return string(e) }
func (e APIKey) GetEnv() string { return os.Getenv(e.String()) }
func (e APIKey) IsValid() bool  { return e.GetEnv() != "" }
func (e APIKey) Validate() error {
	if !e.IsValid() {
		return fmt.Errorf("invalid api key: please export %q", e.String())
	}
	return nil
}

const (
	AIModelReasoningOpenAI        AIModel = "o4-mini"
	AIModelFlagshipOpenAI         AIModel = "gpt-4.1"
	AIModelCostOptimizedOpenAI    AIModel = "gpt-4.1-nano"
	AIModelTTSOpenAI              AIModel = "gpt-4o-mini-tts"
	AIModelTranscriptionOpenAI    AIModel = "gpt-4o-transcribe"
	AIModelReasoningAnthropic     AIModel = "claude-3-7-sonnet-latest"
	AIModelCostOptimizedAnthropic AIModel = "claude-3-5-haiku-latest"
	AIModelTTSElevenLabs          AIModel = "eleven_multilingual_v2"
)

type AIModel string

func (a AIModel) String() string               { return string(a) }
func (a AIModel) MarshalJSON() ([]byte, error) { return json.Marshal(a.String()) }
func (a AIModel) IsValid() bool {
	switch a {
	case AIModelReasoningOpenAI, AIModelFlagshipOpenAI, AIModelCostOptimizedOpenAI,
		AIModelTTSOpenAI, AIModelTranscriptionOpenAI,
		AIModelReasoningAnthropic, AIModelCostOptimizedAnthropic,
		AIModelTTSElevenLabs:
		return true
	default:
		return false
	}
}
func (a AIModel) Validate() error {
	if !a.IsValid() {
		return fmt.Errorf("invalid AI model: %s", a)
	}
	return nil
}

// Temperature controls randomness in model outputs.
type Temperature float32

func (t Temperature) Float32() float32             { return float32(t) }
func (t Temperature) MarshalJSON() ([]byte, error) { return json.Marshal(t.Float32()) }

func (t Temperature) IsValid(m AIModel) bool {
	switch m {
	case AIModelReasoningOpenAI, AIModelFlagshipOpenAI, AIModelCostOptimizedOpenAI:
		return t >= 0 && t <= 2
	case AIModelReasoningAnthropic, AIModelCostOptimizedAnthropic:
		return t >= 0 && t <= 1
	default:
		return false
	}
}

func (t Temperature) Validate(m AIModel) error {
	if !t.IsValid(m) {
		switch m {
		case AIModelReasoningOpenAI, AIModelFlagshipOpenAI, AIModelCostOptimizedOpenAI:
			return fmt.Errorf("invalid temperature for %s: must be 0 <= t <= 2", m)
		case AIModelReasoningAnthropic, AIModelCostOptimizedAnthropic:
			return fmt.Errorf("invalid temperature for %s: must be 0 <= t <= 1", m)
		default:
			return fmt.Errorf("temperature validation not supported for model %s", m)
		}
	}
	return nil
}

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Role string

func (r Role) String() string               { return string(r) }
func (r Role) MarshalJSON() ([]byte, error) { return json.Marshal(r.String()) }
func (r Role) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant:
		return true
	default:
		return false
	}
}
func (r Role) Validate() error {
	if !r.IsValid() {
		return fmt.Errorf("invalid role: %s", r)
	}
	return nil
}

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

func (m Message) IsValid() bool {
	return m.Role.IsValid() && m.Content != ""
}
func (m Message) Validate() error {
	if m.Content == "" {
		return fmt.Errorf("invalid message: message content is empty")
	}
	if !m.Role.IsValid() {
		return m.Role.Validate()
	}
	return nil
}

type Messages []Message

func (ms Messages) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type MaxTokens int

func (mt MaxTokens) Int() int                     { return int(mt) }
func (mt MaxTokens) MarshalJSON() ([]byte, error) { return json.Marshal(mt.Int()) }
func (mt MaxTokens) IsValid() bool                { return mt.Int() >= 1 }
func (mt MaxTokens) Validate() error {
	if !mt.IsValid() {
		return fmt.Errorf("invalid max tokens: must be >= 1")
	}
	return nil
}

type HTTPMethod string

func (hm HTTPMethod) String() string { return string(hm) }
func (hm HTTPMethod) IsValid() bool  { return hm.String() == http.MethodPost }
func (hm HTTPMethod) Validate() error {
	if !hm.IsValid() {
		return errors.New("invalid http method: require POST")
	}
	return nil
}

const (
	OpChatCompletion   Operation = "chat completion"
	OpTTSAudio         Operation = "text-to-speech audio"
	OpSTTTranscription Operation = "speech-to-text transcription"
)

type Operation string

func (o Operation) String() string { return string(o) }
