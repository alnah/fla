package ai

import (
	"errors"
	"strings"
	"testing"
)

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

func TestAIError_Error_Unwrapped(t *testing.T) {
	e := &AIError{
		Operation: OpChatCompletion,
		Provider:  ProviderOpenAI,
		Message:   "something went wrong",
		Wrapped:   nil,
	}
	errStr := e.Error()
	want := "chat completion[openai] error: something went wrong"
	if errStr != want {
		t.Errorf("Error() = %q; want %q", errStr, want)
	}
}

func TestAIError_Error_Wrapped(t *testing.T) {
	orig := errors.New("root cause")
	e := &AIError{
		Operation: OpTTSAudio,
		Provider:  ProviderAnthropic,
		Message:   "tts failed",
		Wrapped:   orig,
	}
	errStr := e.Error()
	if !strings.Contains(errStr, string(OpTTSAudio)) ||
		!strings.Contains(errStr, string(ProviderAnthropic)) ||
		!strings.Contains(errStr, "tts failed") ||
		!strings.Contains(errStr, "root cause") {
		t.Errorf("Error() = %q; missing expected parts", errStr)
	}
}

func TestAIError_Unwrap(t *testing.T) {
	orig := errors.New("underlying")
	e := &AIError{Wrapped: orig}
	if unw := e.Unwrap(); unw != orig {
		t.Errorf("Unwrap() = %v; want %v", unw, orig)
	}
}

func TestNewAIError(t *testing.T) {
	orig := errors.New("orig")
	err := NewAIError(OpSTTTranscription, ProviderElevenLabs, "failed stt", orig)
	var aiErr *AIError
	if !errors.As(err, &aiErr) {
		t.Fatalf("NewAIError did not return *AIError, got %T", err)
	}
	if aiErr.Operation != OpSTTTranscription {
		t.Errorf("Operation = %v; want %v", aiErr.Operation, OpSTTTranscription)
	}
	if aiErr.Provider != ProviderElevenLabs {
		t.Errorf("Provider = %v; want %v", aiErr.Provider, ProviderElevenLabs)
	}
	if aiErr.Message != "failed stt" {
		t.Errorf("Message = %q; want %q", aiErr.Message, "failed stt")
	}
	if aiErr.Wrapped != orig {
		t.Errorf("Wrapped = %v; want %v", aiErr.Wrapped, orig)
	}
}

func TestNewChatError_NewTTSError_NewSTTError(t *testing.T) {
	orig := errors.New("cause")
	e1 := NewChatError(ProviderOpenAI, "chat error", orig)
	e2 := NewTTSError(ProviderAnthropic, "tts error", orig)
	e3 := NewSTTError(ProviderElevenLabs, "stt error", orig)

	var a1, a2, a3 *AIError
	if !errors.As(e1, &a1) || a1.Operation != OpChatCompletion {
		t.Error("NewChatError did not set Operation to OpChatCompletion")
	}
	if !errors.As(e2, &a2) || a2.Operation != OpTTSAudio {
		t.Error("NewTTSError did not set Operation to OpTTSAudio")
	}
	if !errors.As(e3, &a3) || a3.Operation != OpSTTTranscription {
		t.Error("NewSTTError did not set Operation to OpSTTTranscription")
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
