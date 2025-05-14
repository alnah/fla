package aiclient

type TTSClient struct {
	// http fields
	base *BaseClient
	// api fields
	Voice        Voice
	Text         Text
	Instructions Instructions // OpenAI only
	Speed        Speed        // ElevenLabs only
	// provider fields
	UseOpenAI     bool
	UseElevenLabs bool
}
