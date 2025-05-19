package aiclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/transport"
)

func TestTTSClientNew_WithOptions(t *testing.T) {
	testCases := []struct {
		name         string
		provider     provider
		url          url
		apiKey       string
		model        model
		voice        voice
		instructions Instructions
		speed        Speed
		text         Text
	}{
		{
			name:         "OpenAI",
			provider:     ProviderOpenAI,
			url:          URLTTSOpenAI,
			apiKey:       "api-key",
			model:        ModelTTSOpenAI,
			voice:        VoiceOpenAIFemaleAlloy,
			instructions: "test", // openai only
			text:         "test",
		},
		{
			name:     "ElevenLabs",
			provider: ProviderElevenLabs,
			url:      URLTTSElevenLabs,
			apiKey:   "api-key",
			model:    ModelTTSElevenLabs,
			voice:    VoiceElevenLabsFrMaleGuillaume,
			speed:    1, // elevenlabs only
			text:     "text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			log := logger.NewTestLogger()
			httpClient := http.DefaultClient
			httpMethodPost := httpMethod(http.MethodPost)
			tts, err := NewTTSClient(
				WithContext[*TTSClient](ctx),
				WithLogger[*TTSClient](log),
				WithHTTPClient[*TTSClient](httpClient),
				WithHTTPMethod[*TTSClient](httpMethodPost),
				WithProvider[*TTSClient](tc.provider),
				WithURL[*TTSClient](tc.url),
				WithAPIKey[*TTSClient](tc.apiKey),
				WithModel[*TTSClient](tc.model),
				WithVoice(tc.voice),
				WithInstructions(tc.instructions),
				WithSpeed(tc.speed),
				WithText(tc.text),
			)
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
			if tts.base.ctx != ctx {
				t.Errorf("ctx: want %v, got: %v", ctx, tts.base.ctx)
			}
			if tts.base.log != log {
				t.Errorf("logger: want %v, got: %v", log, tts.base.log)
			}
			if tts.base.httpClient != httpClient {
				t.Errorf("http client: want %v, got: %v", httpClient, tts.base.httpClient)
			}
			if tts.base.httpMethod != httpMethodPost {
				t.Errorf("http method: want %v, got: %v", httpMethodPost, tts.base.httpMethod)
			}
			if tts.base.provider != tc.provider {
				t.Errorf("provider: want %v, got: %v", tc.provider, tts.base.provider)
			}
			if tts.base.url != tc.url {
				t.Errorf("url: want %v, got: %v", tc.url, tts.base.url)
			}
			if tts.base.apiKey.String() != tc.apiKey {
				t.Errorf("api key: want %v, got: %v", tc.apiKey, tts.base.apiKey)
			}
			if tts.base.model != tc.model {
				t.Errorf("model: want %v, got: %v", tc.model, tts.base.model)
			}
			if tts.voice != tc.voice {
				t.Errorf("voice: want %v, got: %v", tc.voice, tts.voice)
			}
			if tts.instructions != tc.instructions {
				t.Errorf("instructions: want %v, got: %v", tc.instructions, tts.instructions)
			}
			if tts.speed != tc.speed {
				t.Errorf("speed: want %v, got: %v", tc.speed, tts.speed)
			}
			if tts.text != tc.text {
				t.Errorf("text: want %v, got :%v", tc.text, tts.text)
			}
		})
	}
}

func TestTTSClientNew_Apply_Defaults(t *testing.T) {
	tts, err := NewTTSClient(
		WithProvider[*TTSClient](ProviderOpenAI),
		WithURL[*TTSClient](URLTTSOpenAI),
		WithAPIKey[*TTSClient]("api-key"),
		WithModel[*TTSClient](ModelTTSOpenAI),
		WithVoice(VoiceOpenAIFemaleAlloy),
		WithInstructions("test"),
		WithText("test"),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	testCases := map[string]any{
		"ctx":         tts.base.ctx,
		"logger":      tts.base.log,
		"http client": tts.base.httpClient,
		"http method": tts.base.httpMethod,
	}
	for field, value := range testCases {
		if value == nil {
			t.Errorf("%s: want it set", field)
		}
	}
}

func TestTTSClientNew_Set_ProviderFlag(t *testing.T) {
	testCases := []struct {
		name           string
		provider       provider
		url            url
		apiKey         string
		model          model
		voice          voice
		instructions   Instructions
		speed          Speed
		text           Text
		flagOpenAI     bool
		flagElevenLabs bool
	}{
		{
			name:           "OpenAI",
			provider:       ProviderOpenAI,
			url:            URLTTSOpenAI,
			apiKey:         "api-key",
			model:          ModelTTSOpenAI,
			voice:          VoiceOpenAIFemaleAlloy,
			instructions:   "test",
			text:           "test",
			flagOpenAI:     true,
			flagElevenLabs: false,
		},
		{
			name:           "ElevenLabs",
			provider:       ProviderElevenLabs,
			url:            URLTTSElevenLabs,
			apiKey:         "api-key",
			model:          ModelTTSElevenLabs,
			voice:          VoiceElevenLabsFrFemaleAudrey,
			speed:          1,
			text:           "test",
			flagOpenAI:     false,
			flagElevenLabs: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tts, err := NewTTSClient(
				WithProvider[*TTSClient](tc.provider),
				WithURL[*TTSClient](tc.url),
				WithAPIKey[*TTSClient](tc.apiKey),
				WithModel[*TTSClient](tc.model),
				WithVoice(tc.voice),
				WithInstructions(tc.instructions),
				WithSpeed(tc.speed),
				WithText(tc.text),
			)
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
			switch tc.provider {
			case ProviderOpenAI:
				if tts.useOpenAI != tc.flagOpenAI {
					t.Errorf("openai provider: want flag")
				}
				if tts.useElevenLabs != tc.flagElevenLabs {
					t.Errorf("elevenlabs provider: want no flag")
				}
			case ProviderElevenLabs:
				if tts.useOpenAI != tc.flagOpenAI {
					t.Errorf("openai provider: want no flag")
				}
				if tts.useElevenLabs != tc.flagElevenLabs {
					t.Errorf("elevenlabs provider: want no flag")
				}
			}
		})
	}
}

func TestTTSClientNew_Validate_Fields(t *testing.T) {
	testCases := []struct {
		name       string
		ttsBuilder func() (*TTSClient, error)
		wantError  bool
	}{
		{
			name: "AllRequiredFieldsPassOpenAI",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: false,
		},
		{
			name: "AllRequiredFieldsPassElevenLabs",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			wantError: false,
		},
		{
			name: "AllRequiredAndOptionalFieldsPassOpenAI",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithContext[*TTSClient](context.Background()),
					WithLogger[*TTSClient](logger.NewTestLogger()),
					WithHTTPClient[*TTSClient](http.DefaultClient),
					WithHTTPMethod[*TTSClient](httpMethod(http.MethodPost)),
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: false,
		},
		{
			name: "AllRequiredAndOptionalFieldsPassElevenLabs",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithContext[*TTSClient](context.Background()),
					WithLogger[*TTSClient](logger.NewTestLogger()),
					WithHTTPClient[*TTSClient](http.DefaultClient),
					WithHTTPMethod[*TTSClient](httpMethod(http.MethodPost)),
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			wantError: false,
		},
		{
			name: "RequireProvider",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					// no provider
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireURL",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					// no url
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireAPIKey",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					// no api key
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireModel",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					// no model
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireVoice",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					// no voice
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireInstructionsOpenAIOnly",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					// no instructions
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireSpeedElevenLabsOnly",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					// no speed
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "RequireText",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					// no text
				)
			},
			wantError: true,
		}, {
			name: "InvalidProvider",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](provider("invalid")),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidURL",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](url("invlaid")),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "EmptyAPIKey",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient](""), // empty
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidHTTPMethod",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithHTTPMethod[*TTSClient](httpMethod("INVLAID")),
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidModel",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](model("invalid")),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidVoiceOpenAI",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(voice("invalid")),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidVoiceElevenLabs",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(voice("invalid")),
					WithSpeed(1),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "EmptyInstructionsOpenAIOnly",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions(""), // empty
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "SpeedLt0.7ElevenLabsOnly",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(0.6),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "SpeedGt1.2ElevenLabsOnly",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1.3),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "EmptyText",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText(""), // empty
				)
			},
			wantError: true,
		},
		{
			name: "NilContext",
			ttsBuilder: func() (*TTSClient, error) {
				tts, _ := NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
				// override default context
				tts.base.ctx = nil
				return tts, tts.validate()
			},
			wantError: true,
		},
		{
			name: "NilHTTPClient",
			ttsBuilder: func() (*TTSClient, error) {
				tts, _ := NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
				// override http client
				tts.base.httpClient = nil
				return tts, tts.validate()
			},
			wantError: true,
		},
		{
			name: "NilLogger",
			ttsBuilder: func() (*TTSClient, error) {
				tts, _ := NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
				// override default logger
				tts.base.log = nil
				return tts, tts.validate()
			},
			wantError: true,
		},
		{
			name: "EnsureOneProviderOnly",
			ttsBuilder: func() (*TTSClient, error) {
				tts, _ := NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
				// override provider flags to create conflict
				tts.useOpenAI = true
				tts.useElevenLabs = true
				return tts, tts.validate()
			},
			wantError: true,
		},
		{
			name: "ProviderOpenAIUnmatchingURL",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSElevenLabs), // elevenlabs, want openai
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "ProviderElevenLabsUnmatchingURL",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSOpenAI), // openai, want elevenlabs
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "UnsupportedOpenAIModel",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs), // elevenlabs, want openai
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			wantError: true,
		},
		{
			name: "UnsupportedElevenLabsModel",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI), // openai, want elevenlabs
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ttsBuilder()
			if tc.wantError && err == nil {
				t.Errorf("want error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Errorf("want no error, got %v", err)
			}
		})
	}
}

func TestTTSClient_Audio(t *testing.T) {
	testCases := []struct {
		name            string
		ttsBuilder      func() (*TTSClient, error)
		statusCode      int
		body            *bytes.Buffer
		roundTripperErr error
		wantErr         bool
	}{
		{
			name: "SuccessOpenAI",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			body:       bytes.NewBufferString("irrelevant"),
			statusCode: 200,
			wantErr:    false,
		},
		{
			name: "SuccessElevenLabs",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			body:       bytes.NewBufferString("irrelevant"),
			statusCode: 200,
			wantErr:    false,
		},
		{
			name: "MalformedURL",
			ttsBuilder: func() (*TTSClient, error) {
				tts, _ := NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
				// override url with malformed url
				tts.base.provider = "::::"
				return tts, nil
			},
			body:    bytes.NewBufferString(""),
			wantErr: true,
		},
		{
			name: "NetworkFailed",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			body:            bytes.NewBufferString(""),
			roundTripperErr: errors.New("network failed"),
			wantErr:         true,
		},
		{
			name: "StatusNotOKOpenAI",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			statusCode: 401,
			body: bytes.NewBufferString(
				`{"error": {"message": "incorrect api key provided", "type": "invalid_api_key"}}`,
			),
			wantErr: true,
		},
		{
			name: "StatusNotOKElevenLabs",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			statusCode: 401,
			body: bytes.NewBufferString(
				`{"detail": {"message": "incorrect api key provided", "status": "invalid_api_key"}}`,
			),
			wantErr: true,
		},
		{
			name: "MalformedOpenAIStatusNotOKResponseBody",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderOpenAI),
					WithURL[*TTSClient](URLTTSOpenAI),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSOpenAI),
					WithVoice(VoiceOpenAIFemaleAlloy),
					WithInstructions("test"),
					WithText("test"),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedElevenLabsStatusNotOKResponseBody",
			ttsBuilder: func() (*TTSClient, error) {
				return NewTTSClient(
					WithProvider[*TTSClient](ProviderElevenLabs),
					WithURL[*TTSClient](URLTTSElevenLabs),
					WithAPIKey[*TTSClient]("api-key"),
					WithModel[*TTSClient](ModelTTSElevenLabs),
					WithVoice(VoiceElevenLabsFrFemaleAudrey),
					WithSpeed(1),
					WithText("test"),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tts, _ := tc.ttsBuilder()
			// mock response
			tts.base.httpClient.Transport = transport.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.statusCode, Body: io.NopCloser(tc.body)}, tc.roundTripperErr
			})

			_, err := tts.Audio()
			if tts.base.httpClient.Transport == nil {
				t.Errorf("transport chain: want it set, got nil")
			}
			switch {
			case tc.wantErr:
				if err == nil {
					t.Fatal("want error, got nil")
				}
				t.Log(err.Error())
			default:
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}
			}
		})
	}
}
