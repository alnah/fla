package aiclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/alnah/fla/locale"
	"github.com/alnah/fla/logger"
	"github.com/alnah/fla/pathutil"
	"github.com/alnah/fla/transport"
)

func newTempFile(t testing.TB) pathutil.FilePath {
	t.Helper()

	tempFile, err := os.CreateTemp(".", "test_*.mp3")
	if err != nil {
		t.Fatalf("temp file: want no error, got %v", err)
	}
	t.Cleanup(func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	})

	return pathutil.FilePath(tempFile.Name())
}

func TestSTTClientNew_WithOptions(t *testing.T) {
	ctx := context.Background()
	log := logger.NewTestLogger()
	httpClient := http.DefaultClient
	httpMethodPost := httpMethod(http.MethodPost)
	pvd := ProviderOpenAI
	urlTest := URLSTTOpenAI
	apiKeyTest := APIKeyEnvOpenAI
	modelTest := ModelSTTOpenAI
	filePath := newTempFile(t)
	language := locale.ISO6391("fr")
	stt, err := NewSTTClient(
		WithContext[*STTClient](ctx),
		WithLogger[*STTClient](log),
		WithHTTPClient[*STTClient](httpClient),
		WithHTTPMethod[*STTClient](httpMethodPost),
		WithProvider[*STTClient](pvd),
		WithURL[*STTClient](urlTest),
		WithAPIKey[*STTClient](apiKeyTest),
		WithModel[*STTClient](modelTest),
		WithFilePath(filePath),
		WithLanguage(language),
	)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if stt.base.ctx != ctx {
		t.Errorf("ctx: want %v, got %v", ctx, stt.base.ctx)
	}
	if stt.base.log != log {
		t.Errorf("logger: want %v, got %v", log, stt.base.log)
	}
	if stt.base.httpClient != httpClient {
		t.Errorf("http client: want %v, got %v", httpClient, stt.base.httpClient)
	}
	if stt.base.httpMethod != httpMethodPost {
		t.Errorf("http method: want %v, got %v", httpMethodPost, stt.base.httpMethod)
	}
	if stt.base.provider != pvd {
		t.Errorf("provider: want %v, got %v", pvd, stt.base.provider)
	}
	if stt.base.url != urlTest {
		t.Errorf("url: want %v, got %v", urlTest, stt.base.url)
	}
	if stt.base.apiKey != apiKeyTest {
		t.Errorf("api key: want %v, got %v", apiKeyTest, stt.base.apiKey)
	}
	if stt.base.model != modelTest {
		t.Errorf("model: want %v, got %v", modelTest, stt.base.model)
	}
	if stt.filePath != filePath {
		t.Errorf("file path: want %v, got %v", filePath, stt.filePath)
	}
	if stt.language != language {
		t.Errorf("language: want %v, got %v", locale.ISO6391(language), stt.language)
	}
}

func TestSTTClientNew_Apply_Defaults(t *testing.T) {
	filePath := newTempFile(t)
	tts, err := NewSTTClient(
		WithProvider[*STTClient](ProviderOpenAI),
		WithURL[*STTClient](URLSTTOpenAI),
		WithAPIKey[*STTClient](APIKeyEnvOpenAI),
		WithModel[*STTClient](ModelSTTOpenAI),
		WithLanguage("fr"),
		WithFilePath(filePath),
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

func TestSTTClientNew_Set_ProviderFlag(t *testing.T) {
	testCases := []struct {
		name           string
		provider       provider
		url            url
		apiKey         apiKey
		model          model
		language       locale.ISO6391
		filePath       pathutil.FilePath
		flagOpenAI     bool
		flagElevenLabs bool
	}{
		{
			name:           "OpenAI",
			provider:       ProviderOpenAI,
			url:            URLSTTOpenAI,
			apiKey:         APIKeyEnvOpenAI,
			model:          ModelSTTOpenAI,
			language:       "fr",
			filePath:       newTempFile(t),
			flagOpenAI:     true,
			flagElevenLabs: false,
		},
		{
			name:           "ElevenLabs",
			provider:       ProviderElevenLabs,
			url:            URLSTTElevenLabs,
			apiKey:         APIKeyEnvElevenLabs,
			model:          ModelSTTElevenLabs,
			language:       "fr",
			filePath:       newTempFile(t),
			flagOpenAI:     false,
			flagElevenLabs: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tts, err := NewSTTClient(
				WithProvider[*STTClient](tc.provider),
				WithURL[*STTClient](tc.url),
				WithAPIKey[*STTClient](tc.apiKey),
				WithModel[*STTClient](tc.model),
				WithLanguage(tc.language),
				WithFilePath(tc.filePath),
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

func TestSTTNew_Validate_Fields(t *testing.T) {
	testCases := []struct {
		name       string
		sttBuilder func() (*STTClient, error)
		wantError  bool
	}{
		{
			name: "AllRequiredFieldsPass",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: false,
		},
		{
			name: "AllRequiredAndOptionalFieldsPass",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithContext[*STTClient](context.Background()),
					WithLogger[*STTClient](logger.NewTestLogger()),
					WithHTTPClient[*STTClient](http.DefaultClient),
					WithHTTPMethod[*STTClient](httpMethod(http.MethodPost)),
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: false,
		},
		{
			name: "RequireProvider",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					// no provider
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "RequireURL",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					// no url
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "RequireAPIKey",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					// no api key
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "RequireModel",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					// no model
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "RequireLanguage",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					// no language
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "RequireFilePath",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					// require file path
				)
			},
			wantError: true,
		},
		{
			name: "InvalidHttpMethod",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithHTTPMethod[*STTClient](httpMethod("invalid")),
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidProvider",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](provider("invalid")),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidURL",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](url("invalid")),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidAPIKey",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](apiKey("NON_EXISTING_API_KEY")),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidModel",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](model("invalid")),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidLanguage",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("invalid"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
		{
			name: "InvalidFilePath",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath("./non/existing/file/path"),
				)
			},
			wantError: true,
		},
		{
			name: "NilContext",
			sttBuilder: func() (*STTClient, error) {
				stt, _ := NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
				// override context
				stt.base.ctx = nil
				_, err := stt.validate()
				return stt, err
			},
			wantError: true,
		},
		{
			name: "NilLogger",
			sttBuilder: func() (*STTClient, error) {
				stt, _ := NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
				// override logger
				stt.base.log = nil
				_, err := stt.validate()
				return stt, err
			},
			wantError: true,
		},
		{
			name: "NilHTTPClient",
			sttBuilder: func() (*STTClient, error) {
				stt, _ := NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
				// override http client
				stt.base.httpClient = nil
				_, err := stt.validate()
				return stt, err
			},
			wantError: true,
		},
		{
			name: "EnsureOneProviderOnly",
			sttBuilder: func() (*STTClient, error) {
				stt, _ := NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
				// override provider flags to create conflict
				stt.useOpenAI = true
				stt.useElevenLabs = true
				_, err := stt.validate()
				return stt, err
			},
			wantError: true,
		}, {
			name: "ProviderOpenAIUnmatchingURL",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTElevenLabs), // elevenlabs, want openai
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		}, {
			name: "ProviderElevenLabsUnmatchingURL",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderElevenLabs),
					WithURL[*STTClient](URLSTTOpenAI), // openai, want elevenlabs
					WithAPIKey[*STTClient](APIKeyEnvElevenLabs),
					WithModel[*STTClient](ModelSTTElevenLabs),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		}, {
			name: "UnsupportedOpenAIModel",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTElevenLabs), // elevenlabs, want openai
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		}, {
			name: "UnsupportedElevenLabsModel",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderElevenLabs),
					WithURL[*STTClient](URLSTTElevenLabs),
					WithAPIKey[*STTClient](APIKeyEnvElevenLabs),
					WithModel[*STTClient](ModelSTTOpenAI), // openai, want elevenlabs
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					_, err := tc.sttBuilder()
					if tc.wantError && err == nil {
						t.Errorf("want error, got nil")
					}
					if !tc.wantError && err != nil {
						t.Errorf("want no error, got %v", err)
					}
				})
			}
		})
	}
}

func TestSTTClient_Transcrip(t *testing.T) {
	testCases := []struct {
		name            string
		sttBuilder      func() (*STTClient, error)
		statusCode      int
		body            *bytes.Buffer
		roundTripperErr error
		want            string
		wantErr         bool
	}{
		{
			name: "SuccessOpenAI",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"text": "test"}`),
			want:       "test",
			wantErr:    false,
		},
		{
			name: "SuccessElevenLabs",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderElevenLabs),
					WithURL[*STTClient](URLSTTElevenLabs),
					WithAPIKey[*STTClient](APIKeyEnvElevenLabs),
					WithModel[*STTClient](ModelSTTElevenLabs),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"text": "test"}`),
			want:       "test",
			wantErr:    false,
		},
		{
			name: "MalformedURL",
			sttBuilder: func() (*STTClient, error) {
				chat, _ := NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
				// override url with malformed url
				chat.base.url = "::::"
				return chat, nil
			},
			body:    bytes.NewBufferString(""),
			wantErr: true,
		},
		{
			name: "NetworkFailed",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			body:            bytes.NewBufferString(""),
			roundTripperErr: errors.New("network error"),
			wantErr:         true,
		},
		{
			name: "StatusNotOKOpenAI",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
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
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderElevenLabs),
					WithURL[*STTClient](URLSTTElevenLabs),
					WithAPIKey[*STTClient](APIKeyEnvElevenLabs),
					WithModel[*STTClient](ModelSTTElevenLabs),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 401,
			body: bytes.NewBufferString(
				`{"detail": {"message": "incorrect api key provided", "status": "invalid_api_key"}}`,
			),
			wantErr: true,
		},
		{
			name: "MalformedResponseBody",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			body:    bytes.NewBufferString("{]]invalid[[}"),
			wantErr: true,
		},
		{
			name: "MalformedOpenAIStatusOKResponseBody",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedElevenLabsStatusOKResponseBody",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderElevenLabs),
					WithURL[*STTClient](URLSTTElevenLabs),
					WithAPIKey[*STTClient](APIKeyEnvElevenLabs),
					WithModel[*STTClient](ModelSTTElevenLabs),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 200,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedOpenAIStatusNotOKResponseBody",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderOpenAI),
					WithURL[*STTClient](URLSTTOpenAI),
					WithAPIKey[*STTClient](APIKeyEnvOpenAI),
					WithModel[*STTClient](ModelSTTOpenAI),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
		{
			name: "MalformedElevenLabsStatusNotOKResponseBody",
			sttBuilder: func() (*STTClient, error) {
				return NewSTTClient(
					WithProvider[*STTClient](ProviderElevenLabs),
					WithURL[*STTClient](URLSTTElevenLabs),
					WithAPIKey[*STTClient](APIKeyEnvElevenLabs),
					WithModel[*STTClient](ModelSTTElevenLabs),
					WithLanguage("fr"),
					WithFilePath(newTempFile(t)),
				)
			},
			statusCode: 401,
			body:       bytes.NewBufferString(`{"malformed:""}`),
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tts, _ := tc.sttBuilder()
			// mock response
			tts.base.httpClient.Transport = transport.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.statusCode, Body: io.NopCloser(tc.body)}, tc.roundTripperErr
			})

			transript, err := tts.Transcript()
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
				if transript.Content() != tc.want {
					t.Errorf("content: want %s, got %s", tc.want, transript.Content())
				}
			}
		})
	}
}
