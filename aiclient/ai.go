package aiclient

// Chat defines behaviors to perform a chat completion.
type Chat interface {
	SetSystem(instructions string) *Chat
	AddMessage(role Role, content string) *Chat
	Completion() (Completion, error)
}

// TTS defines behaviors to synthetize a text-to-speech audio.
type TTS interface {
	Audio() ([]byte, error)
}

// STT defines behaviors to perform a speech-to-text transcription.
type STT interface {
	Transcript() (Transcription, error)
}

// Env vars holding API keys for AI providers.
const (
	EnvOpenAIAPIKey     string = "OPENAI_API_KEY"     // #nosec G101
	EnvAnthropicAPIKey  string = "ANTHROPIC_API_KEY"  // #nosec G101
	EnvElevenLabsAPIKey string = "ELEVENLABS_API_KEY" // #nosec G101
)

// Api represents an API name service providing AI resources: OpenAI, Anthropic and ElevenLabs.
type Api string

const (
	APIOpenAI     Api = "openai"
	APIAnthropic  Api = "anthropic"
	APIElevenLabs Api = "elevenlabs"
)

// Op represents a chat completion, a text-to-speech audio, or a speech-to-text transcription.
type Op string

const (
	OpChatCompletion   Op = "chat completion"
	OpTTSAudio         Op = "text-to-speech audio"
	OpSTTTranscription Op = "speech-to-text transcription"
)

// Role represents the unisque system prompt-level, the user's messages, or the assistant's messages.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)
