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
	ProviderOpenAI     provider = "openai"
	ProviderAnthropic  provider = "anthropic"
	ProviderElevenLabs provider = "elevenlabs"
)

type provider string

func (p provider) String() string { return string(p) }
func (p provider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderElevenLabs:
		return true
	default:
		return false
	}
}
func (p provider) Validate() error {
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
	URLChatOpenAI           url = "https://api.openai.com/v1/chat/completions"
	URLChatAnthropic        url = "https://api.anthropic.com/v1/messages"
	URLSpeechOpenAI         url = "https://api.openai.com/v1/audio/speech"
	URLSpeechElevenLabs     url = "https://api.elevenlabs.io/v1/text-to-speech"
	URLTranscriptOpenAI     url = "https://api.openai.com/v1/audio/transcriptions"
	URLTranscriptElevenLabs url = "https://api.elevenlabs.io/v1/speech-to-text"
)

type url string

func (u url) String() string { return string(u) }
func (u url) IsValid() bool {
	switch u {
	case URLChatOpenAI, URLChatAnthropic,
		URLSpeechOpenAI, URLSpeechElevenLabs,
		URLTranscriptOpenAI:
		return true
	default:
		return false
	}
}
func (u url) Validate() error {
	if u.String() == "" {
		return fmt.Errorf("invalid url: can't be empty")
	}
	if !u.IsValid() {
		return fmt.Errorf("invalid url: %s, please use correct gateway and endpoint", u)
	}
	return nil
}

const (
	EnvOpenAIAPIKey     apiKey = "OPENAI_API_KEY"     // #nosec G101
	EnvAnthropicAPIKey  apiKey = "ANTHROPIC_API_KEY"  // #nosec G101
	EnvElevenLabsAPIKey apiKey = "ELEVENLABS_API_KEY" // #nosec G101
)

type apiKey string

func (e apiKey) String() string { return string(e) }
func (e apiKey) GetEnv() string { return os.Getenv(e.String()) }
func (e apiKey) IsValid() bool  { return e.GetEnv() != "" }
func (e apiKey) Validate() error {
	if e.String() == "" {
		return fmt.Errorf("invalid api key: can't be empty")
	}
	if !e.IsValid() {
		return fmt.Errorf("invalid api key: please export %q env var", e.String())
	}
	return nil
}

type httpMethod string

func (hm httpMethod) String() string { return string(hm) }
func (hm httpMethod) IsValid() bool  { return hm.String() == http.MethodPost }
func (hm httpMethod) Validate() error {
	if !hm.IsValid() {
		return errors.New("invalid http method: require POST")
	}
	return nil
}

const (
	OpChatCompletion operation = "chat completion"
	OpSpeechAudio    operation = "audio speech"
	OpTranscript     operation = "transcript"
)

type operation string

func (o operation) String() string { return string(o) }

const (
	AIModelReasoningOpenAI        aiModel = "o4-mini"
	AIModelFlagshipOpenAI         aiModel = "gpt-4.1"
	AIModelCostOptimizedOpenAI    aiModel = "gpt-4.1-nano"
	AIModelSpeechOpenAI           aiModel = "gpt-4o-mini-tts"
	AIModelTranscriptionOpenAI    aiModel = "gpt-4o-transcribe"
	AIModelReasoningAnthropic     aiModel = "claude-3-7-sonnet-latest"
	AIModelCostOptimizedAnthropic aiModel = "claude-3-5-haiku-latest"
	AIModelTTSElevenLabs          aiModel = "eleven_multilingual_v2"
)

type aiModel string

func (a aiModel) String() string               { return string(a) }
func (a aiModel) MarshalJSON() ([]byte, error) { return json.Marshal(a.String()) }
func (a aiModel) IsValid() bool {
	switch a {
	case AIModelReasoningOpenAI, AIModelFlagshipOpenAI, AIModelCostOptimizedOpenAI,
		AIModelSpeechOpenAI, AIModelTranscriptionOpenAI,
		AIModelReasoningAnthropic, AIModelCostOptimizedAnthropic,
		AIModelTTSElevenLabs:
		return true
	default:
		return false
	}
}
func (a aiModel) Validate() error {
	if a.String() == "" {
		return fmt.Errorf("invalid ai model: can't be empty")
	}
	if !a.IsValid() {
		available := strings.Join([]string{
			AIModelReasoningOpenAI.String(),
			AIModelFlagshipOpenAI.String(),
			AIModelCostOptimizedOpenAI.String(),
			AIModelSpeechOpenAI.String(),
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

func (t Temperature) IsValid(m aiModel) bool {
	switch m {
	case AIModelReasoningOpenAI, AIModelFlagshipOpenAI, AIModelCostOptimizedOpenAI:
		return t >= 0 && t <= 2
	case AIModelReasoningAnthropic, AIModelCostOptimizedAnthropic:
		return t >= 0 && t <= 1
	default:
		return false
	}
}

func (t Temperature) Validate(m aiModel) error {
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
	RoleSystem    role = "system"
	RoleUser      role = "user"
	RoleAssistant role = "assistant"
)

type role string

func (r role) String() string               { return string(r) }
func (r role) MarshalJSON() ([]byte, error) { return json.Marshal(r.String()) }
func (r role) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant:
		return true
	default:
		return false
	}
}
func (r role) Validate() error {
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
	Role    role   `json:"role"`
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

type voice string

// OpenAI
const (
	VoiceOpenAIFemaleAlloy   voice = "alloy"
	VoiceOpenAIMaleAsh       voice = "ash"
	VoiceOpenAIMaleBallad    voice = "ballad"
	VoiceOpenAIFemaleCoral   voice = "coral"
	VoiceOpenAIMaleEcho      voice = "echo"
	VoiceOpenAINeutralFable  voice = "fable"
	VoiceOpenAIMaleOnyx      voice = "onyx"
	VoiceOpenAIFemaleNova    voice = "nova"
	VoiceOpenAIFemaleSage    voice = "sage"
	VoiceOpenAIFemaleShimmer voice = "shimmer"
	VoiceOpenAIMaleVerse     voice = "verse"
)

// ElevenLabs
const (
	// FR
	VoiceElevenLabsFrMaleNicolas   voice = "aQROLel5sQbj1vuIVi6B"
	VoiceElevenLabsFrMaleGuillaume voice = "ohItIVrXTBI80RrUECOD"
	VoiceElevenLabsFrFemaleAudrey  voice = "McVZB9hVxVSk3Equu8EH"
	// PT
	VoiceElevenLabsPtMaleMarcelo voice = "bJrNspxJVFovUxNBQ0wh"
	VoiceElevenLabsPtMaleSamuel  voice = "ETf5cmpNIbpSiXmBaR2m"
	VoiceElevenLabsPtFemaleBia   voice = "Eyspt3SYhZzXd1Jd3J8O"
)

func (v voice) String() string               { return string(v) }
func (v voice) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }
func (v voice) IsValid(p provider) bool {
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

func (v voice) Validate(p provider) error {
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
			availableMap := map[voice]string{
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
func (i Instructions) IsValid(p provider) bool {
	if p == ProviderOpenAI && i.String() == "" {
		return false
	}
	return true
}
func (i Instructions) Validate(p provider) error {
	if !i.IsValid(p) {
		return errors.New("invalid instructions: can't be empty")
	}
	return nil
}

type Speed float32

func (s Speed) Float32() float32             { return float32(s) }
func (s Speed) MarshalJSON() ([]byte, error) { return json.Marshal(s.Float32()) }
func (s Speed) IsValid(p provider) bool {
	if p == ProviderElevenLabs && (s.Float32() < 0.7 || s.Float32() > 1.2) {
		return false
	}
	return true
}

func (s Speed) Validate(p provider) error {
	if !s.IsValid(p) {
		return fmt.Errorf("invalid speed: must be 0.7 <= s <= 1.2")
	}
	return nil
}
