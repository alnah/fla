package ai

import "testing"

/********* Interface *********/

type enumlike interface {
	String() string
	IsValid() bool
}

/********* Helpers *********/

func buildAssertString[T enumlike](name string) func(t testing.TB, got, want string) {
	return func(t testing.TB, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func buildAssertIsValid[T enumlike](name string) func(t testing.TB, enum T) {
	return func(t testing.TB, enum T) {
		t.Helper()
		if !enum.IsValid() {
			t.Errorf("%s %s should be valid", name, enum.String())
		}
	}
}

/********* Unit tests *********/

func TestRoleString(t *testing.T) {
	assertString := buildAssertString[Role]("role")
	assertString(t, RoleSystem.String(), "system")
	assertString(t, RoleUser.String(), "user")
	assertString(t, RoleAssistant.String(), "assistant")
}

func TestRoleIsValid(t *testing.T) {
	assertIsValid := buildAssertIsValid[Role]("role")
	assertIsValid(t, RoleSystem)
	assertIsValid(t, RoleUser)
	assertIsValid(t, RoleAssistant)
}

func TestProviderString(t *testing.T) {
	assertString := buildAssertString[Provider]("provider")
	assertString(t, ProviderOpenAI.String(), "openai")
	assertString(t, ProviderAnthropic.String(), "anthropic")
	assertString(t, ProviderElevenLabs.String(), "elevenlabs")
}

func TestProviderIsValid(t *testing.T) {
	assertIsValid := buildAssertIsValid[Provider]("provider")
	assertIsValid(t, ProviderOpenAI)
	assertIsValid(t, ProviderAnthropic)
	assertIsValid(t, ProviderElevenLabs)
}

func TestOperationString(t *testing.T) {
	assertString := buildAssertString[Operation]("operation")
	assertString(t, OpChatCompletion.String(), "chat completion")
	assertString(t, OpTTSAudio.String(), "text-to-speech audio")
	assertString(t, OpSTTTranscription.String(), "speech-to-text transcription")
}

func TestOperationIsValid(t *testing.T) {
	assertIsValid := buildAssertIsValid[Operation]("operation")
	assertIsValid(t, OpChatCompletion)
	assertIsValid(t, OpTTSAudio)
	assertIsValid(t, OpSTTTranscription)
}

func TestCompletionContent(t *testing.T) {
	tests := []struct {
		name      string
		openai    OpenAICompletion
		anthropic AnthropicCompletion
		want      string
	}{
		{
			name: "OpenAI wins",
			openai: OpenAICompletion{Message: struct {
				Content string "json:\"content,omitempty\""
			}{"test"}},
			anthropic: AnthropicCompletion{},
			want:      "test",
		},
		{
			name:      "Anthropic fallback",
			openai:    OpenAICompletion{},
			anthropic: AnthropicCompletion{Content: "test"},
			want:      "test",
		},
		{
			name:      "Empty",
			openai:    OpenAICompletion{},
			anthropic: AnthropicCompletion{},
			want:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := Completion{
				OpenAI:    []OpenAICompletion{tc.openai},
				Anthropic: []AnthropicCompletion{tc.anthropic},
			}
			if got := c.Content(); got != tc.want {
				t.Errorf("completion: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTranscriptionContent(t *testing.T) {
	tr := Transcription{Text: "test"}
	if got := tr.Content(); got != "test" {
		t.Errorf("transcription: %q; want %q", got, "test")
	}
}
