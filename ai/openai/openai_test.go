package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/clog"
)

/********* Helpers *********/

type roundTripperTest func(req *http.Request) (*http.Response, error)

func (f roundTripperTest) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type marker string

const key marker = "marker"

/********* Chat Completion Unit Tests *********/

func TestChat_WithOption_Pattern(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")
	msg := []ai.Message{{Role: ai.RoleUser, Content: "test"}}

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if got := req.Context().Value(key); got != "value" {
			t.Error("http request: want context inheritance from chat")
		}
		body := `{"content": [{"text": "ok", "type":"" }], "id": "1"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}

	chat := NewChat(
		WithModel[*Chat](ModelReasoning),
		WithLogger[*Chat](log),
		WithHTTPClient[*Chat](hc),
		WithContext[*Chat](ctx),
		WithAPIKey[*Chat]("test-api-key"),
		WithURL[*Chat]("http://test.com"),
		WithMessages(msg),
		WithMaxTokens(1234),
		WithTemperature(0.5),
	)

	if chat.Model.String() != ModelReasoning.String() {
		t.Errorf("with model: want %v, got %v", ModelReasoning.String(), chat.Model.String())
	}
	if chat.log != log {
		t.Errorf("with logger: want %v, got %v", log, chat.log)
	}
	if chat.hc != hc {
		t.Errorf("with HTTP client: want %v, got %v", hc, chat.hc)
	}
	if chat.ctx != ctx {
		t.Errorf("with context: want %v, got %v", ctx, chat.ctx)
	}
	if chat.apiKey != "test-api-key" {
		t.Errorf("with api key: want %q, got %q", "test-api-key", chat.apiKey)
	}
	if chat.url != "http://test.com" {
		t.Errorf("with url: want %q, got %q", "http://test.com", chat.url)
	}
	if chat.MaxTokens != 1234 {
		t.Errorf("with max tokens: want %d, got %d", 1234, chat.MaxTokens)
	}
	if chat.Temperature != 0.5 {
		t.Errorf("with temperature: want %v, got %v", 0.5, chat.Temperature)
	}
	if len(chat.Messages) != 1 || chat.Messages[0] != msg[0] {
		t.Errorf("with messages: want %+v, got %+v", msg, chat.Messages)
	}
}
func TestChat_SetSystem_OnlyOnce(t *testing.T) {
	chat := NewChat().
		SetSystem("want no change").
		SetSystem("got change")

	if chat.Messages[0].Content == "got change" {
		t.Error("chat system: want no change, got change")
	}
}

func TestChat_SetSystemAndAddMessage_FluentPattern(t *testing.T) {
	chat := NewChat().
		SetSystem("system-msg").
		AddMessage(ai.RoleUser, "user-msg").
		AddMessage(ai.RoleAssistant, "assistant-msg")

	if got := len(chat.Messages); got != 3 {
		t.Errorf("chat messages length: want %d, got %d", 3, got)
	}
	got1, got2, got3 := chat.Messages[0].Content, chat.Messages[1].Content, chat.Messages[2].Content
	want1, want2, want3 := "system-msg", "user-msg", "assistant-msg"
	if got1 != want1 || got2 != want2 || got3 != want3 {
		t.Errorf("messages: want [%s, %s, %s], got [%s, %s, %s]", want1, want2, want3, got1, got2, got3)
	}
}

func TestChat_Completion_EmptyMessages(t *testing.T) {
	var want *ai.AIError
	if _, err := NewChat().Completion(); !errors.As(err, &want) {
		t.Errorf("empty messages: want ai error, got %T", err)
	}
}

func TestChat_Completion_InvalidURL(t *testing.T) {
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey[*Chat]("test"),
		WithURL[*Chat]("::::"),
	)

	var want *url.Error
	if _, err := chat.Completion(); !errors.As(err, &want) {
		t.Errorf("invalid request: want error as url error, got %T", err)
	}
}

func TestChat_Completion_InvalidJSONResponse(t *testing.T) {
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`[invalid}]`))}, nil
	})}

	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey[*Chat]("test-api-key"),
		WithURL[*Chat]("https://test.com"),
		WithHTTPClient[*Chat](hc),
	)

	var want *json.SyntaxError
	if _, err := chat.Completion(); !errors.As(err, &want) {
		t.Errorf("invalid json: want error as json syntax error, got %T", err)
	}
}

func TestChat_Completion_ContextTimeout(t *testing.T) {
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(2 * time.Millisecond):
			body := `{"content": [{"text": "ok", "type":"" }], "id": "1"}`
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
	})}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	chat := NewChat(
		WithContext[*Chat](ctx),
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey[*Chat]("test-api-key"),
		WithURL[*Chat]("https://test.com"),
		WithHTTPClient[*Chat](hc),
	)

	var want *ai.AIError
	_, err := chat.Completion()
	if !errors.As(err, &want) {
		t.Error("context timeout: want ai error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("context timeout: ai error must unwrap context deadline exceeded")
	}
}

func TestChat_Completion_HTTPClientError(t *testing.T) {
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network failed")
	})}

	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey[*Chat]("test-api-key"),
		WithURL[*Chat]("https://test.com"),
		WithHTTPClient[*Chat](hc),
	)

	var want *ai.AIError
	if _, err := chat.Completion(); !errors.As(err, &want) {
		t.Fatalf("http client failed: want ai error, got %T", err)
	}
}

func TestChat_Completion_HTTPErrorStatus(t *testing.T) {
	body, _ := json.Marshal(&ai.HTTPError{Message: "message", Type: "type"})
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewReader(body))}, nil
	})}

	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey[*Chat]("test-api-key"),
		WithURL[*Chat]("https://test.com"),
		WithHTTPClient[*Chat](hc),
	)

	var want *ai.HTTPError
	if _, err := chat.Completion(); !errors.As(err, &want) {
		t.Errorf("http error status: want http error, got %T", err)
	}
}

func TestChat_Completion_Success(t *testing.T) {
	body := `{"content": [{"text":"hello", "type":""}], "id": "123"}`
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}

	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey[*Chat]("test-api-key"),
		WithURL[*Chat]("https://test.com"),
		WithHTTPClient[*Chat](hc),
	)

	completion, err := chat.Completion()
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if got := completion.Content(); got != "hello" {
		t.Errorf("completion content: \"hello\", got %q", got)
	}
}

func TestChat_Completion_ContentEmpty(t *testing.T) {
	var completion ai.Completion
	if completion.Content() != "" {
		t.Errorf("want empty content, got %q", completion.Content())
	}
}

/********* TTS Audio Unit Tests *********/

func TestTTS_WithOption_Pattern(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if got := req.Context().Value(key); got != "value" {
			t.Error("http request: want context inheritance from chat")
		}
		body := `{"content": [{"text": "ok", "type":"" }], "id": "1"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}

	tts := NewTTS(
		WithModel[*TTS](ModelTextToSpeech),
		WithLogger[*TTS](log),
		WithHTTPClient[*TTS](hc),
		WithContext[*TTS](ctx),
		WithAPIKey[*TTS]("test-api-key"),
		WithURL[*TTS]("http://test.com"),
		WithInput("test"),
		WithVoice(VoiceFemaleShimmer),
	)

	if tts.Model.String() != ModelTextToSpeech.String() {
		t.Errorf("with model: want %v, got %v", ModelTextToSpeech.String(), tts.Model.String())
	}
	if tts.log != log {
		t.Errorf("with logger: want %v, got %v", log, tts.log)
	}
	if tts.hc != hc {
		t.Errorf("with HTTP client: want %v, got %v", hc, tts.hc)
	}
	if tts.ctx != ctx {
		t.Errorf("with context: want %v, got %v", ctx, tts.ctx)
	}
	if tts.apiKey != "test-api-key" {
		t.Errorf("with api key: want %q, got %q", "test-api-key", tts.apiKey)
	}
	if tts.url != "http://test.com" {
		t.Errorf("with url: want %q, got %q", "http://test.com", tts.url)
	}
	if tts.Input != "test" {
		t.Errorf("with input: want %q, got %q", "test", tts.Input)
	}
	if tts.Voice.String() != VoiceFemaleShimmer.String() {
		t.Errorf("with voice: want %q, got %q", VoiceFemaleShimmer.String(), tts.Voice.String())
	}
}

func TestTTS_Audio_EmptyInput(t *testing.T) {
	var want *ai.AIError
	if _, err := NewTTS(WithInput("")).Audio(); !errors.As(err, &want) {
		t.Errorf("input empty: want ai error, got %T", err)
	}
}

func TestTTS_Audio_InvalidURL(t *testing.T) {
	tts := NewTTS(
		WithInput("test"),
		WithAPIKey[*TTS]("test-api-key"),
		WithURL[*TTS]("::::"),
	)

	var want *url.Error
	if _, err := tts.Audio(); !errors.As(err, &want) {
		t.Errorf("invalid request: want url error, got %T", err)
	}
}

func TestTTS_Audio_ContextTimeout(t *testing.T) {
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(2 * time.Millisecond):
			body := `{"content": [{"text": "ok", "type":"" }], "id": "1"}`
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
	})}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	chat := NewTTS(
		WithContext[*TTS](ctx),
		WithInput("test"),
		WithAPIKey[*TTS]("test-api-key"),
		WithURL[*TTS]("https://test.com"),
		WithHTTPClient[*TTS](hc),
	)

	var want *ai.AIError
	_, err := chat.Audio()
	if !errors.As(err, &want) {
		t.Error("context timeout: want ai error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("context timeout: ai error must unwrap context deadline exceeded")
	}
}

func TestAudio_HTTPClientError(t *testing.T) {
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network fail")
	})}

	tts := NewTTS(
		WithInput("test"),
		WithAPIKey[*TTS]("test-api-key"),
		WithURL[*TTS]("https://test.com"),
		WithHTTPClient[*TTS](hc),
	)

	var want *ai.AIError
	if _, err := tts.Audio(); !errors.As(err, &want) {
		t.Errorf("http client error: want ai error, but got %T", err)
	}
}

func TestAudio_HTTPStatusError(t *testing.T) {
	body, _ := json.Marshal(&ai.HTTPError{Message: "message", Type: "type"})
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewReader(body))}, nil
	})}

	chat := NewTTS(
		WithInput("test"),
		WithAPIKey[*TTS]("test-api-key"),
		WithURL[*TTS]("https://test.com"),
		WithHTTPClient[*TTS](hc),
	)

	var want *ai.HTTPError
	if _, err := chat.Audio(); !errors.As(err, &want) {
		t.Errorf("http error status: want http error, got %T", err)
	}
}

func TestAudio_Success(t *testing.T) {
	data := []byte("audio-bytes")
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})}

	tts := NewTTS(
		WithInput("hey"),
		WithAPIKey[*TTS]("test-api-key"),
		WithURL[*TTS]("http://u/"),
		WithHTTPClient[*TTS](hc),
	)

	audio, err := tts.Audio()
	if err != nil {
		t.Fatalf("want no error: got %v", err)
	}
	if !bytes.Equal(audio, data) {
		t.Errorf("audio bytes: want %v, got %v", audio, data)
	}
}

/********* STT Transcription Unit Test *********/

func TestSTT_WithOption_Pattern(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")
	f, err := os.CreateTemp("", "stt-test-*.wav")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if got := req.Context().Value(key); got != "value" {
			t.Error("http request: want context inheritance from stt")
		}
		// renvoyer une réponse valide pour ne pas bloquer l'initialisation
		body := `{"text": "ok"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}

	stt := NewSTT(
		WithModel[*STT](ModelSpeechToText),
		WithLogger[*STT](log),
		WithHTTPClient[*STT](hc),
		WithContext[*STT](ctx),
		WithAPIKey[*STT]("test-api-key"),
		WithURL[*STT]("http://test.com"),
		WithFile(f),
		WithFilePath(f.Name()),
		WithLanguage("fr"),
	)

	if stt.Model.String() != ModelSpeechToText.String() {
		t.Errorf("with model: want %v, got %v", ModelSpeechToText, stt.Model)
	}
	if stt.log != log {
		t.Errorf("with logger: want %v, got %v", log, stt.log)
	}
	if stt.hc != hc {
		t.Errorf("with http client: want %v, got %v", hc, stt.hc)
	}
	if stt.ctx != ctx {
		t.Errorf("with context: want %v, got %v", ctx, stt.ctx)
	}
	if stt.apiKey != "test-api-key" {
		t.Errorf("with api key: want %q, got %q", "test-api-key", stt.apiKey)
	}
	if stt.url != "http://test.com" {
		t.Errorf("with url: want %q, got %q", "http://test.com", stt.url)
	}
	if stt.File != f {
		t.Errorf("with file: want %v, got %v", f, stt.File)
	}
	if stt.FilePath != f.Name() {
		t.Errorf("with filepath: want %q, got %q", f.Name(), stt.FilePath)
	}
	if stt.Language != "fr" {
		t.Errorf("with language: want %q, got %q", "fr", stt.Language)
	}
}

func TestSTT_Transcript_MissingFile(t *testing.T) {
	stt := NewSTT(WithFilePath("p"), WithLanguage("en"))
	var want *ai.AIError
	if _, err := stt.Transcript(); !errors.As(err, &want) {
		t.Errorf("missing file: want ai error, got %T", err)
	}
}

func TestSTT_Transcript_MissingFilePath(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	stt := NewSTT(WithFile(f), WithLanguage("en"))
	var want *ai.AIError
	if _, err := stt.Transcript(); !errors.As(err, &want) {
		t.Errorf("missing filepath: want ai error, got %T", err)
	}
}

func TestSTT_Transcript_MissingLanguage(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	stt := NewSTT(WithFile(f), WithFilePath(f.Name()))
	var want *ai.AIError
	if _, err := stt.Transcript(); !errors.As(err, &want) {
		t.Errorf("missing language: want ai error, got %T", err)
	}
}

func TestSTT_Transcript_InvalidURL(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	stt := NewSTT(
		WithFile(f),
		WithFilePath(f.Name()),
		WithLanguage("en"),
		WithAPIKey[*STT]("test-api-key"),
		WithURL[*STT]("::::"),
	)

	var want *url.Error
	if _, err := stt.Transcript(); !errors.As(err, &want) {
		t.Errorf("invalid request: want url error, got %T", err)
	}
}

func TestSTT_Transcript_ContextTimeout(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(2 * time.Millisecond):
			body := `{"text":"ok"}`
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
	})}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	stt := NewSTT(
		WithContext[*STT](ctx),
		WithFile(f),
		WithFilePath(f.Name()),
		WithLanguage("en"),
		WithAPIKey[*STT]("test-api-key"),
		WithURL[*STT]("https://test.com"),
		WithHTTPClient[*STT](hc),
	)

	var want *ai.AIError
	_, err := stt.Transcript()
	if !errors.As(err, &want) {
		t.Error("context timeout: want ai error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("context timeout: ai error must unwrap context deadline exceeded")
	}
}

func TestSTT_Transcript_HTTPClientError(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network fail")
	})}

	stt := NewSTT(
		WithFile(f),
		WithFilePath(f.Name()),
		WithLanguage("en"),
		WithAPIKey[*STT]("test-api-key"),
		WithURL[*STT]("https://test.com"),
		WithHTTPClient[*STT](hc),
	)

	var want *ai.AIError
	if _, err := stt.Transcript(); !errors.As(err, &want) {
		t.Errorf("http client error: want ai error, got %T", err)
	}
}

func TestSTT_Transcript_HTTPStatusError(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	body, _ := json.Marshal(&ai.HTTPError{Message: "msg", Type: "type"})
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewReader(body))}, nil
	})}

	stt := NewSTT(
		WithFile(f),
		WithFilePath(f.Name()),
		WithLanguage("en"),
		WithAPIKey[*STT]("test-api-key"),
		WithURL[*STT]("https://test.com"),
		WithHTTPClient[*STT](hc),
	)

	var want *ai.HTTPError
	if _, err := stt.Transcript(); !errors.As(err, &want) {
		t.Errorf("http status error: want HTTPError, got %T", err)
	}
}

func TestSTT_Transcript_Success(t *testing.T) {
	f, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(f.Name())

	resp := ai.Transcription{Text: "test"}
	data, _ := json.Marshal(resp)
	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})}

	stt := NewSTT(
		WithFile(f),
		WithFilePath(f.Name()),
		WithLanguage("fr"),
		WithAPIKey[*STT]("test-api-key"),
		WithURL[*STT]("https://test.com"),
		WithHTTPClient[*STT](hc),
	)

	out, err := stt.Transcript()
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if out.Text != "test" {
		t.Errorf("transcription text: want %q, got %q", "test", out.Text)
	}
}

func TestSTT_Transcript_InvalidJSONResponse(t *testing.T) {
	tmp, _ := os.CreateTemp("", "stt-*.wav")
	defer os.Remove(tmp.Name())

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`[invalid json]`)),
		}, nil
	})}

	stt := NewSTT(
		WithFile(tmp),
		WithFilePath(tmp.Name()),
		WithLanguage("fr"),
		WithAPIKey[*STT]("test"),
		WithURL[*STT]("https://test.com"),
		WithHTTPClient[*STT](hc),
	)

	var want *json.SyntaxError
	_, err := stt.Transcript()
	if !errors.As(err, &want) {
		t.Errorf("invalid json: want *json.SyntaxError, got %T", err)
	}
}
