package aiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	ProviderOpenAI     provider = "openai"
	ProviderAnthropic  provider = "anthropic"
	ProviderElevenLabs provider = "elevenlabs"
)

// provider identifies which AI service (OpenAI, Anthropic, ElevenLabs) is in use,
// enabling provider-specific behavior downstream.
type provider string

// String returns the provider’s name.
func (p provider) String() string { return string(p) }

// IsValid checks whether the provider is one of the supported constants.
func (p provider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderElevenLabs:
		return true
	default:
		return false
	}
}

// Validate ensures a provider has been set and is supported,
// returning an error otherwise.
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
	URLChatOpenAI    url = "https://api.openai.com/v1/chat/completions"
	URLChatAnthropic url = "https://api.anthropic.com/v1/messages"
	URLTTSOpenAI     url = "https://api.openai.com/v1/audio/speech"
	URLTTSElevenLabs url = "https://api.elevenlabs.io/v1/text-to-speech"
	URLSTTOpenAI     url = "https://api.openai.com/v1/audio/transcriptions"
	URLSTTElevenLabs url = "https://api.elevenlabs.io/v1/speech-to-text"
)

// url represents an API endpoint URL for a given provider,
// enabling compile-time validation against known endpoints.
type url string

// String returns the endpoint URL.
func (u url) String() string { return string(u) }

// IsValid checks whether the URL matches one of the known endpoints.
func (u url) IsValid() bool {
	switch u {
	case URLChatOpenAI, URLChatAnthropic,
		URLTTSOpenAI, URLTTSElevenLabs,
		URLSTTOpenAI, URLSTTElevenLabs:
		return true
	default:
		return false
	}
}

// Validate ensures a URL has been set and is supported,
// returning an error otherwise.
func (u url) Validate() error {
	if u.String() == "" {
		return fmt.Errorf("invalid url: can't be empty")
	}
	if !u.IsValid() {
		return fmt.Errorf("invalid url: %s, please use correct gateway and endpoint", u)
	}
	return nil
}

// internal type satisfying
type rawAPIKey string

func (r rawAPIKey) String() string { return string(r) }
func (r rawAPIKey) IsValid() bool  { return r != "" }
func (r rawAPIKey) Validate() error {
	if r == "" {
		return errors.New("invalid api key: can't be empty")
	}
	return nil
}

// httpMethod specifies which HTTP verb to use; currently only POST is supported.
type httpMethod string

// String returns the HTTP method.
func (hm httpMethod) String() string { return string(hm) }

// IsValid checks whether the method is POST.
func (hm httpMethod) IsValid() bool { return hm.String() == http.MethodPost }

// Validate ensures the method is POST, returning an error otherwise.
func (hm httpMethod) Validate() error {
	if !hm.IsValid() {
		return errors.New("invalid http method: require POST")
	}
	return nil
}

const (
	opChatCompletion operation = "chat completion"
	opTextToSpeech   operation = "text-to-speech audio"
	opSpeechToText   operation = "speech-to-text transcript"
)

// operation describes a high-level API action for error reporting.
type operation string

// String returns the operation’s human-readable name.
func (o operation) String() string { return string(o) }

const (
	// chat completion
	ModelReasoningOpenAI    model = "o4-mini"
	ModelFlagshipOpenAI     model = "gpt-4.1"
	ModelCheapOpenAI        model = "gpt-4.1-nano"
	ModelReasoningAnthropic model = "claude-3-7-sonnet-latest"
	ModelCheapAnthropic     model = "claude-3-5-haiku-latest"

	// text-to-speech audio
	ModelTTSOpenAI     model = "gpt-4o-mini-tts"
	ModelTTSElevenLabs model = "eleven_multilingual_v2"

	// speech-to-text transcript
	ModelSTTOpenAI     model = "gpt-4o-transcribe"
	ModelSTTElevenLabs model = "scribe_v1"
)

// model enumerates the supported AI model identifiers,
// enforcing valid selection at build time.
type model string

// String returns the model’s identifier.
func (a model) String() string { return string(a) }

// MarshalJSON serializes the model into JSON.
func (a model) MarshalJSON() ([]byte, error) { return json.Marshal(a.String()) }

// IsValid checks whether the model is one of the supported set.
func (a model) IsValid() bool {
	switch a {
	case ModelReasoningOpenAI, ModelFlagshipOpenAI, ModelCheapOpenAI,
		ModelTTSOpenAI, ModelSTTOpenAI,
		ModelReasoningAnthropic, ModelCheapAnthropic,
		ModelTTSElevenLabs, ModelSTTElevenLabs:
		return true
	default:
		return false
	}
}

// Validate ensures a model has been set and is supported,
// returning an error otherwise.
func (a model) Validate() error {
	if a.String() == "" {
		return fmt.Errorf("invalid ai model: can't be empty")
	}
	if !a.IsValid() {
		available := strings.Join([]string{
			ModelReasoningOpenAI.String(),
			ModelFlagshipOpenAI.String(),
			ModelCheapOpenAI.String(),
			ModelTTSOpenAI.String(),
			ModelSTTOpenAI.String(),
			ModelReasoningAnthropic.String(),
			ModelCheapAnthropic.String(),
			ModelTTSElevenLabs.String(),
		}, ", ")
		return fmt.Errorf("invalid ai model: %s, available models: %s", a.String(), available)
	}
	return nil
}

// Temperature controls the randomness of model outputs;
// valid ranges differ between providers and models.
type Temperature float32

// Float32 returns the numeric value.
func (t Temperature) Float32() float32 { return float32(t) }

// MarshalJSON serializes the temperature into JSON.
func (t Temperature) MarshalJSON() ([]byte, error) { return json.Marshal(t.Float32()) }

// IsValid checks whether the temperature is within the allowed range for a model.
func (t Temperature) IsValid(m model) bool {
	switch m {
	case ModelReasoningOpenAI, ModelFlagshipOpenAI, ModelCheapOpenAI:
		return t >= 0 && t <= 2
	case ModelReasoningAnthropic, ModelCheapAnthropic:
		return t >= 0 && t <= 1
	default:
		return false
	}
}

// Validate ensures the temperature is valid for a given model,
// returning an error otherwise.
func (t Temperature) Validate(m model) error {
	if !t.IsValid(m) {
		switch m {
		case ModelReasoningOpenAI, ModelFlagshipOpenAI, ModelCheapOpenAI:
			return fmt.Errorf("invalid temperature for %s: must be 0 <= t <= 2", m)
		case ModelReasoningAnthropic, ModelCheapAnthropic:
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

// role identifies the sender of a message in a chat:
// system, user or assistant.
type role string

// String returns the role’s name.
func (r role) String() string { return string(r) }

// MarshalJSON serializes the role into JSON.
func (r role) MarshalJSON() ([]byte, error) { return json.Marshal(r.String()) }

// IsValid checks whether the role is one of the allowed values.
func (r role) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant:
		return true
	default:
		return false
	}
}

// Validate ensures a role is recognized, returning an error otherwise.
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

// Message binds a role to its textual content in a chat exchange.
type Message struct {
	Role    role   `json:"role"`
	Content string `json:"content"`
}

// IsValid checks that a Message has both a valid role and non-empty content.
func (m Message) IsValid() bool {
	return m.Role.IsValid() && m.Content != ""
}

// Validate ensures the message is well-formed, returning an error otherwise.
func (m Message) Validate() error {
	if !m.Role.IsValid() {
		return m.Role.Validate()
	}
	if !m.IsValid() {
		return fmt.Errorf("invalid message: can't be empty")
	}
	return nil
}

// Messages is a slice of Message, validating each entry for correctness.
type Messages []Message

// Validate checks every Message in the slice, returning the first error encountered.
func (ms Messages) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// MaxTokens caps how many tokens the model may generate in a completion.
type MaxTokens int

// Int returns the value as a native int.
func (mt MaxTokens) Int() int { return int(mt) }

// MarshalJSON serializes the max-tokens into JSON.
func (mt MaxTokens) MarshalJSON() ([]byte, error) { return json.Marshal(mt.Int()) }

// IsValid checks that MaxTokens is at least 1.
func (mt MaxTokens) IsValid() bool { return mt.Int() >= 1 }

// Validate ensures the token limit is valid, returning an error otherwise.
func (mt MaxTokens) Validate() error {
	if !mt.IsValid() {
		return fmt.Errorf("invalid max tokens: must be >= 1")
	}
	return nil
}

// Text wraps input strings for TTS and ensures they are non-empty.
type Text string

// String returns the raw text.
func (t Text) String() string { return string(t) }

// MarshalJSON serializes the text into JSON.
func (t Text) MarshalJSON() ([]byte, error) { return json.Marshal(t.String()) }

// IsValid checks that the string is not empty.
func (t Text) IsValid() bool { return t != "" }

// Validate ensures the text is non-empty, returning an error otherwise.
func (t Text) Validate() error {
	if !t.IsValid() {
		return errors.New("invalid text: can't be empty")
	}
	return nil
}

// voice identifies which TTS voice to use, varying by provider.
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

// String returns the voice’s ID.
func (v voice) String() string { return string(v) }

// MarshalJSON serializes the voice into JSON.
func (v voice) MarshalJSON() ([]byte, error) { return json.Marshal(v.String()) }

// IsValid checks whether the voice is supported by a given provider.
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

// Validate ensures the voice is valid for its provider,
// returning an error otherwise.
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

// Instructions conveys additional guidance to the TTS engine;
// required for OpenAI’s TTS.
type Instructions string

// String returns the instruction text.
func (i Instructions) String() string { return string(i) }

// MarshalJSON serializes the instructions into JSON.
func (i Instructions) MarshalJSON() ([]byte, error) { return json.Marshal(i) }

// IsValid checks whether instructions meet provider requirements.
func (i Instructions) IsValid(p provider) bool {
	if p == ProviderOpenAI && i.String() == "" {
		return false
	}
	return true
}

// Validate ensures instructions are non-empty when required,
// returning an error otherwise.
func (i Instructions) Validate(p provider) error {
	if !i.IsValid(p) {
		return errors.New("invalid instructions: can't be empty")
	}
	return nil
}

// Speed controls playback rate for ElevenLabs TTS,
// constrained to a provider-defined range.
type Speed float32

// Float32 returns the numeric value.
func (s Speed) Float32() float32 { return float32(s) }

// MarshalJSON serializes the speed into JSON.
func (s Speed) MarshalJSON() ([]byte, error) { return json.Marshal(s.Float32()) }

// IsValid checks whether the speed is within the allowed range.
func (s Speed) IsValid(p provider) bool {
	if p == ProviderElevenLabs && (s.Float32() < 0.7 || s.Float32() > 1.2) {
		return false
	}
	return true
}

// Validate ensures the speed is within bounds,
// returning an error otherwise.
func (s Speed) Validate(p provider) error {
	if !s.IsValid(p) {
		return fmt.Errorf("invalid speed: must be 0.7 <= s <= 1.2")
	}
	return nil
}
