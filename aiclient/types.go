package aiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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
		return fmt.Errorf("invalid provider: can't be empty")
	}
	if !p.IsValid() {
		providers := strings.Join([]string{ProviderOpenAI.String(), ProviderAnthropic.String()}, ", ")
		return fmt.Errorf("invalid provider: %s, available providers: %s", p.String(), providers)
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
	if u.String() == "" {
		return fmt.Errorf("invalid url: can't be empty")
	}
	if !u.IsValid() {
		return fmt.Errorf("invalid url: %s, please use correct gateway and endpoint", u)
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
	if e.String() == "" {
		return fmt.Errorf("invalid api key: can't be empty")
	}
	if !e.IsValid() {
		return fmt.Errorf("invalid api key: please export %q env var", e.String())
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
	if a.String() == "" {
		return fmt.Errorf("invalid ai model: can't be empty")
	}
	if !a.IsValid() {
		available := strings.Join([]string{
			AIModelReasoningOpenAI.String(),
			AIModelFlagshipOpenAI.String(),
			AIModelCostOptimizedOpenAI.String(),
			AIModelTTSOpenAI.String(),
			AIModelTranscriptionOpenAI.String(),
			AIModelReasoningAnthropic.String(),
			AIModelCostOptimizedAnthropic.String(),
			AIModelTTSElevenLabs.String(),
		}, ", ")
		return fmt.Errorf("invalid ai model: %s, available models: %s", a.String(), available)
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
		available := strings.Join([]string{
			RoleSystem.String(),
			RoleUser.String(),
			RoleAssistant.String(),
		}, ", ")
		return fmt.Errorf("invalid role: %s, available roles: %s", r.String(), available)
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
	if !m.Role.IsValid() {
		return m.Role.Validate()
	}
	if !m.IsValid() {
		return fmt.Errorf("invalid message: can't be empty")
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

type Text string

func (t Text) String() string               { return string(t) }
func (t Text) MarshalJSON() ([]byte, error) { return json.Marshal(t.String()) }
func (t Text) IsValid() bool                { return t != "" }
func (t Text) Validate() error {
	if !t.IsValid() {
		return errors.New("invalid text: can't be empty")
	}
	return nil
}

type Voice string

// OpenAI
const (
	VoiceOpenAIFemaleAlloy   Voice = "alloy"
	VoiceOpenAIMaleAsh       Voice = "ash"
	VoiceOpenAIMaleBallad    Voice = "ballad"
	VoiceOpenAIFemaleCoral   Voice = "coral"
	VoiceOpenAIMaleEcho      Voice = "echo"
	VoiceOpenAINeutralFable  Voice = "fable"
	VoiceOpenAIMaleOnyx      Voice = "onyx"
	VoiceOpenAIFemaleNova    Voice = "nova"
	VoiceOpenAIFemaleSage    Voice = "sage"
	VoiceOpenAIFemaleShimmer Voice = "shimmer"
	VoiceOpenAIMaleVerse     Voice = "verse"
)

// ElevenLabs
const (
	// FR
	VoiceElevenLabsFrMaleNicolas   Voice = "aQROLel5sQbj1vuIVi6B"
	VoiceElevenLabsFrMaleGuillaume Voice = "ohItIVrXTBI80RrUECOD"
	VoiceElevenLabsFrFemaleAudrey  Voice = "McVZB9hVxVSk3Equu8EH"
	// PT
	VoiceElevenLabsPtMaleMarcelo Voice = "bJrNspxJVFovUxNBQ0wh"
	VoiceElevenLabsPtMaleSamuel  Voice = "ETf5cmpNIbpSiXmBaR2m"
	VoiceElevenLabsPtFemaleBia   Voice = "Eyspt3SYhZzXd1Jd3J8O"
)

func (v Voice) String() string               { return string(v) }
func (v Voice) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }
func (v Voice) IsValid(p Provider) bool {
	switch p {
	case ProviderOpenAI:
		return v == VoiceOpenAIFemaleAlloy ||
			v == VoiceOpenAIMaleAsh ||
			v == VoiceOpenAIMaleBallad ||
			v == VoiceOpenAIFemaleCoral ||
			v == VoiceOpenAIMaleEcho ||
			v == VoiceOpenAINeutralFable ||
			v == VoiceOpenAIMaleOnyx ||
			v == VoiceOpenAIFemaleNova ||
			v == VoiceOpenAIFemaleSage ||
			v == VoiceOpenAIFemaleShimmer ||
			v == VoiceOpenAIMaleVerse
	case ProviderElevenLabs:
		return v == VoiceElevenLabsFrMaleNicolas ||
			v == VoiceElevenLabsFrMaleGuillaume ||
			v == VoiceElevenLabsFrFemaleAudrey ||
			v == VoiceElevenLabsPtMaleMarcelo ||
			v == VoiceElevenLabsPtMaleSamuel ||
			v == VoiceElevenLabsPtFemaleBia
	default:
		return false
	}
}

func (v Voice) Validate(p Provider) error {
	if v.String() == "" {
		return errors.New("invalid voice: can't be empty")
	}
	if !v.IsValid(p) {
		switch p {
		case ProviderOpenAI:
			available := strings.Join([]string{
				VoiceOpenAIMaleAsh.String(),
				VoiceOpenAIMaleBallad.String(),
				VoiceOpenAIFemaleCoral.String(),
				VoiceOpenAIMaleEcho.String(),
				VoiceOpenAINeutralFable.String(),
				VoiceOpenAIMaleOnyx.String(),
				VoiceOpenAIFemaleNova.String(),
				VoiceOpenAIFemaleSage.String(),
				VoiceOpenAIFemaleShimmer.String(),
				VoiceOpenAIMaleVerse.String(),
			}, ", ")
			return fmt.Errorf("invalid voice: %s, available voices: %s", v.String(), available)
		case ProviderElevenLabs:
			availableMap := map[Voice]string{
				VoiceElevenLabsFrFemaleAudrey:  "audrey (fr)",
				VoiceElevenLabsFrMaleGuillaume: "guillaume (fr)",
				VoiceElevenLabsFrMaleNicolas:   "nicolas (fr)",
				VoiceElevenLabsPtFemaleBia:     "bia (pt)",
				VoiceElevenLabsPtMaleMarcelo:   "marcelo (pt)",
				VoiceElevenLabsPtMaleSamuel:    "samuel (pt)",
			}
			var availableSlice []string
			for _, name := range availableMap {
				availableSlice = append(availableSlice, name)
			}
			available := strings.Join(availableSlice, ", ")
			return fmt.Errorf("invalid voice: %s, available voices: %s", v.String(), available)
		}
	}
	return nil
}

type Instructions string

func (i Instructions) String() string               { return string(i) }
func (i Instructions) MarshalJSON() ([]byte, error) { return json.Marshal(i) }
func (i Instructions) IsValid(p Provider) bool {
	if p == ProviderOpenAI && i.String() == "" {
		return false
	}
	return true
}
func (i Instructions) Validate(p Provider) error {
	if !i.IsValid(p) {
		return errors.New("invalid instructions: can't be empty")
	}
	return nil
}

type Speed float32

func (s Speed) Float32() float32             { return float32(s) }
func (s Speed) MarshalJSON() ([]byte, error) { return json.Marshal(s.Float32()) }
func (s Speed) IsValid(p Provider) bool {
	if p == ProviderElevenLabs && (s.Float32() < 0.7 || s.Float32() > 1.2) {
		return false
	}
	return true
}

func (s Speed) Validate(p Provider) error {
	if !s.IsValid(p) {
		return fmt.Errorf("invalid speed: must be 0.7 <= s <= 1.2")
	}
	return nil
}
