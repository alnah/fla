package ai

import "testing"

func TestRoleString(t *testing.T) {
	if got, want := RoleSystem.String(), "system"; got != want {
		t.Errorf("RoleSystem.String() = %q; want %q", got, want)
	}
	if got, want := RoleUser.String(), "user"; got != want {
		t.Errorf("RoleUser.String() = %q; want %q", got, want)
	}
	if got, want := RoleAssistant.String(), "assistant"; got != want {
		t.Errorf("RoleAssistant.String() = %q; want %q", got, want)
	}
}

func TestRoleIsValid(t *testing.T) {
	if !RoleSystem.IsValid() {
		t.Error("RoleSystem should be valid")
	}
	if !RoleUser.IsValid() {
		t.Error("RoleUser should be valid")
	}
	if !RoleAssistant.IsValid() {
		t.Error("RoleAssistant should be valid")
	}
	if Role("bogus").IsValid() {
		t.Error(`Role("bogus") should be invalid`)
	}
}

func TestProviderString(t *testing.T) {
	if got, want := ProviderOpenAI.String(), "openai"; got != want {
		t.Errorf("ProviderOpenAI.String() = %q; want %q", got, want)
	}
	if got, want := ProviderAnthropic.String(), "anthropic"; got != want {
		t.Errorf("ProviderAnthropic.String() = %q; want %q", got, want)
	}
	if got, want := ProviderElevenLabs.String(), "elevenlabs"; got != want {
		t.Errorf("ProviderElevenLabs.String() = %q; want %q", got, want)
	}
}

func TestProviderIsValid(t *testing.T) {
	if !ProviderOpenAI.IsValid() {
		t.Error("ProviderOpenAI should be valid")
	}
	if !ProviderAnthropic.IsValid() {
		t.Error("ProviderAnthropic should be valid")
	}
	if !ProviderElevenLabs.IsValid() {
		t.Error("ProviderElevenLabs should be valid")
	}
	if Provider("bogus").IsValid() {
		t.Error(`Provider("bogus") should be invalid`)
	}
}

func TestOperationString(t *testing.T) {
	if got, want := OpChatCompletion.String(), "chat completion"; got != want {
		t.Errorf("OpChatCompletion.String() = %q; want %q", got, want)
	}
	if got, want := OpTTSAudio.String(), "text-to-speech audio"; got != want {
		t.Errorf("OpTTSAudio.String() = %q; want %q", got, want)
	}
	if got, want := OpSTTTranscription.String(), "speech-to-text transcription"; got != want {
		t.Errorf("OpSTTTranscription.String() = %q; want %q", got, want)
	}
}

func TestOperationIsValid(t *testing.T) {
	if !OpChatCompletion.IsValid() {
		t.Error("OpChatCompletion should be valid")
	}
	if !OpTTSAudio.IsValid() {
		t.Error("OpTTSAudio should be valid")
	}
	if !OpSTTTranscription.IsValid() {
		t.Error("OpSTTTranscription should be valid")
	}
	if Operation("bogus").IsValid() {
		t.Error(`Operation("bogus") should be invalid`)
	}
}

func TestCompletionContent_OpenAI(t *testing.T) {
	var oc OpenAICompletion
	oc.Message.Content = "hello from openai"
	c := Completion{
		OpenAI:    []OpenAICompletion{oc},
		Anthropic: []AnthropicCompletion{{Content: "hello from anthropic"}},
	}
	if got := c.Content(); got != "hello from openai" {
		t.Errorf("Content() = %q; want %q", got, "hello from openai")
	}
}

func TestCompletionContent_Anthropic(t *testing.T) {
	var emptyOC OpenAICompletion
	c := Completion{
		OpenAI:    []OpenAICompletion{emptyOC},
		Anthropic: []AnthropicCompletion{{Content: "hello anthropic"}},
	}
	if got := c.Content(); got != "hello anthropic" {
		t.Errorf("Content() = %q; want %q", got, "hello anthropic")
	}
}

func TestCompletionContent_Empty(t *testing.T) {
	c := Completion{}
	if got := c.Content(); got != "" {
		t.Errorf("Content() = %q; want empty string", got)
	}
}

func TestTranscriptionContent(t *testing.T) {
	tr := Transcription{Text: "transcribed text"}
	if got := tr.Content(); got != "transcribed text" {
		t.Errorf("Content() = %q; want %q", got, "transcribed text")
	}
}
