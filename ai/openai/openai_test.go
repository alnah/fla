package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/ai/transport"
	"github.com/alnah/fla/clog"
)

type roundTripperTest func(req *http.Request) (*http.Response, error)

func (f roundTripperTest) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

/********* Option pattern *********/

type marker string

const key marker = "marker"

func TestChatGenericOptions(t *testing.T) {
	// setup common fixtures
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")
	// use a client that verifies context inheritance
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if req.Context().Value(key) != "value" {
			t.Error("HTTP request did not inherit context from Chat")
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})}
	msg := ai.Message{Role: ai.RoleUser, Content: "hello"}

	chat := NewChat(
		WithModel[*Chat](ModelFlagship),
		WithLogger[*Chat](log),
		WithHTTPClient[*Chat](client),
		WithContext[*Chat](ctx),
		WithAPIKey[*Chat]("testkey"),
		WithURL[*Chat]("http://example.com"),
		WithMaxCompletionTokens(77),
		WithTemperature(0.33),
		WithMessages([]ai.Message{msg}),
	)

	if chat.Base.Model != ModelFlagship {
		t.Errorf("WithModel: got %v, want %v", chat.Base.Model, ModelFlagship)
	}
	if chat.Base.log != log {
		t.Error("WithLogger did not set logger on Base")
	}
	if chat.Base.hc != client {
		t.Error("WithHTTPClient did not set HTTP client on Base")
	}
	if chat.Base.ctx != ctx {
		t.Error("WithContext did not set context on Base")
	}
	if chat.Base.apiKey != "testkey" {
		t.Errorf("WithAPIKey: got %q, want testkey", chat.Base.apiKey)
	}
	if chat.Base.url != "http://example.com" {
		t.Errorf("WithURL: got %q, want http://example.com", chat.Base.url)
	}
	if chat.MaxCompletionTokens != 77 {
		t.Errorf("WithMaxCompletionTokens: got %d, want 77", chat.MaxCompletionTokens)
	}
	if chat.Temperature != 0.33 {
		t.Errorf("WithTemperature: got %v, want 0.33", chat.Temperature)
	}
	if len(chat.Messages) != 1 || chat.Messages[0] != msg {
		t.Errorf("WithMessages: got %+v, want [%+v]", chat.Messages, msg)
	}

	// trigger completion to exercise transport chain
	chat.Completion()
}

func TestTTSGenericOptions(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")
	client := &http.Client{}

	tts := NewTTS(
		WithModel[*TTS](ModelTextToSpeech),
		WithLogger[*TTS](log),
		WithHTTPClient[*TTS](client),
		WithContext[*TTS](ctx),
		WithAPIKey[*TTS]("k"),
		WithURL[*TTS]("u"),
		WithInput("inp"),
		WithVoice(VoiceMaleEcho),
		WithInstructions("inst"),
	)

	if tts.Base.Model != ModelTextToSpeech {
		t.Errorf("WithModel: got %v, want %v", tts.Base.Model, ModelTextToSpeech)
	}
	if tts.Base.log != log {
		t.Error("WithLogger did not set logger on TTS.Base")
	}
	if tts.Base.hc != client {
		t.Error("WithHTTPClient did not set HTTP client on TTS.Base")
	}
	if tts.Base.ctx != ctx {
		t.Error("WithContext did not set context on TTS.Base")
	}
	if tts.Base.apiKey != "k" {
		t.Errorf("WithAPIKey: got %q, want k", tts.Base.apiKey)
	}
	if tts.Base.url != "u" {
		t.Errorf("WithURL: got %q, want u", tts.Base.url)
	}
	if tts.Input != "inp" {
		t.Errorf("WithInput: got %q, want inp", tts.Input)
	}
	if tts.Voice != VoiceMaleEcho {
		t.Errorf("WithVoice: got %v, want %v", tts.Voice, VoiceMaleEcho)
	}
	if tts.Instructions != "inst" {
		t.Errorf("WithInstructions: got %q, want inst", tts.Instructions)
	}
}

func TestSTTGenericOptions(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")
	client := &http.Client{}
	tmp, err := os.CreateTemp("", "stt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	stt := NewSTT(
		WithModel[*STT](ModelSpeechToText),
		WithLogger[*STT](log),
		WithHTTPClient[*STT](client),
		WithContext[*STT](ctx),
		WithAPIKey[*STT]("kk"),
		WithURL[*STT]("uu"),
		WithFile(tmp),
		WithFilePath(tmp.Name()),
		WithLanguage("fr"),
	)

	if stt.Base.Model != ModelSpeechToText {
		t.Errorf("WithModel: got %v, want %v", stt.Base.Model, ModelSpeechToText)
	}
	if stt.Base.log != log {
		t.Error("WithLogger did not set logger on STT.Base")
	}
	if stt.Base.hc != client {
		t.Error("WithHTTPClient did not set HTTP client on STT.Base")
	}
	if stt.Base.ctx != ctx {
		t.Error("WithContext did not set context on STT.Base")
	}
	if stt.Base.apiKey != "kk" {
		t.Errorf("WithAPIKey: got %q, want kk", stt.Base.apiKey)
	}
	if stt.Base.url != "uu" {
		t.Errorf("WithURL: got %q, want uu", stt.Base.url)
	}
	if stt.File != tmp {
		t.Errorf("WithFile: got %v, want %v", stt.File, tmp)
	}
	if stt.FilePath != tmp.Name() {
		t.Errorf("WithFilePath: got %q, want %q", stt.FilePath, tmp.Name())
	}
	if stt.Language != "fr" {
		t.Errorf("WithLanguage: got %q, want fr", stt.Language)
	}
}

/********* Chat.Completion *********/

func TestChatCompletion_NoMessages(t *testing.T) {
	chat := NewChat()
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
	if !strings.Contains(err.Error(), "messages required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatAddMessageFluent(t *testing.T) {
	chat := NewChat()
	chat.AddMessage(ai.RoleUser, "one").AddMessage(ai.RoleAssistant, "two")

	if len(chat.Messages) != 2 {
		t.Fatalf("AddMessage appended %d messages; want 2", len(chat.Messages))
	}
	msg1, msg2 := chat.Messages[0], chat.Messages[1]
	if msg1.Role != ai.RoleUser || msg1.Content != "one" {
		t.Errorf("first message = %+v; want RoleUser/one", msg1)
	}
	if msg2.Role != ai.RoleAssistant || msg2.Content != "two" {
		t.Errorf("second message = %+v; want RoleAssistant/two", msg2)
	}
}

func TestChatCompletion_HTTPClientError(t *testing.T) {
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network fail")
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "hello"}}),
		WithAPIKey[*Chat]("testkey"),
		WithURL[*Chat]("http://example.com"),
		WithHTTPClient[*Chat](client),
	)
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected error for HTTP client failure")
	}
	if !strings.Contains(err.Error(), "failed to send HTTP request") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatCompletion_HTTPErrorStatus(t *testing.T) {
	apiErr := transport.APIError{}
	apiErr.Error.Type = "type"
	apiErr.Error.Code = "code"
	apiErr.Error.Message = "msg"
	b, _ := json.Marshal(apiErr)
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewReader(b)),
		}, nil
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "hello"}}),
		WithAPIKey[*Chat]("testkey"),
		WithURL[*Chat]("http://example.com"),
		WithHTTPClient[*Chat](client),
	)
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected HTTP error for non-200 status")
	}
	he, ok := err.(*transport.HTTPError)
	if !ok {
		t.Fatalf("expected *transport.HTTPError, got %T", err)
	}
	if he.Status != 400 || he.Type != "type" || he.Code != "code" || he.Message != "msg" {
		t.Errorf("unexpected HTTPError: %+v", he)
	}
}

func TestChatCompletion_InvalidJSON(t *testing.T) {
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("invalid json")),
		}, nil
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "hello"}}),
		WithAPIKey[*Chat]("testkey"),
		WithURL[*Chat]("http://example.com"),
		WithHTTPClient[*Chat](client),
	)
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatCompletion_Success(t *testing.T) {
	respBody := `{"choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]}`
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(respBody)),
		}, nil
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "hello"}}),
		WithAPIKey[*Chat]("testkey"),
		WithURL[*Chat]("http://example.com"),
		WithHTTPClient[*Chat](client),
	)
	res, err := chat.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content() != "hi" {
		t.Errorf("expected content 'hi', got %q", res.Content())
	}
}

func TestCompletion_ContentEmpty(t *testing.T) {
	var res completion
	if res.Content() != "" {
		t.Errorf("expected empty content, got %q", res.Content())
	}
}

/********* TTS.Audio *********/

func TestTTSAudio_NoInput(t *testing.T) {
	tts := NewTTS()
	_, err := tts.Audio()
	if err == nil {
		t.Fatal("expected error for missing input")
	}
	if !strings.Contains(err.Error(), "text input required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTTSAudio_HTTPError(t *testing.T) {
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network fail")
	})}
	tts := NewTTS(
		WithInput("hello"),
		WithAPIKey[*TTS]("testkey"),
		WithURL[*TTS]("http://example.com"),
		WithHTTPClient[*TTS](client),
	)
	_, err := tts.Audio()
	if err == nil {
		t.Fatal("expected error for HTTP client failure")
	}
	if !strings.Contains(err.Error(), "failed to send HTTP request") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTTSAudio_Success(t *testing.T) {
	data := []byte("audio data")
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})}
	tts := NewTTS(
		WithInput("hello"),
		WithAPIKey[*TTS]("testkey"),
		WithURL[*TTS]("http://example.com"),
		WithHTTPClient[*TTS](client),
	)
	out, err := tts.Audio()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Errorf("expected %v, got %v", data, out)
	}
}

/********* STT.Transcript *********/

func TestSTTTranscript_FileRequired(t *testing.T) {
	stt := NewSTT()
	_, err := stt.Transcript()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "file required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSTTTranscript_FilepathRequired(t *testing.T) {
	tmp, err := os.CreateTemp("", "stt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	stt := NewSTT(WithFile(tmp))
	_, err = stt.Transcript()
	if err == nil {
		t.Fatal("expected error for missing filepath")
	}
	if !strings.Contains(err.Error(), "filepath required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSTTTranscript_LanguageRequired(t *testing.T) {
	tmp, err := os.CreateTemp("", "stt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	stt := NewSTT(
		WithFile(tmp),
		WithFilePath(tmp.Name()),
	)
	_, err = stt.Transcript()
	if err == nil {
		t.Fatal("expected error for missing language")
	}
	if !strings.Contains(err.Error(), "source language required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSTTTranscript_Success(t *testing.T) {
	tmp, err := os.CreateTemp("", "stt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString("audio data")
	tmp.Seek(0, 0)

	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		body := `{"text":"transcribed"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	stt := NewSTT(
		WithFile(tmp),
		WithFilePath(tmp.Name()),
		WithLanguage("en"),
		WithHTTPClient[*STT](client),
	)
	res, err := stt.Transcript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content() != "transcribed" {
		t.Errorf("expected transcript 'transcribed', got %q", res.Content())
	}
}

/********* Integration tests *********/

func TestIntegrationChatCompletion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// setup a test http server to mimic the openai chat completion endpoint
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		// decode request to verify payload
		var body Chat
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(body.Messages) == 0 {
			t.Fatalf("no messages in request")
		}
		// respond with a fabricated completion
		w.Header().Set("Content-Type", "application/json")
		type choiceWrapper struct {
			Message ai.Message `json:"message"`
		}
		resp := struct {
			Choices []choiceWrapper `json:"choices"`
		}{
			Choices: []choiceWrapper{{Message: ai.Message{Role: ai.RoleAssistant, Content: "integration reply"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := srv.Client()
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "ping"}}),
		WithAPIKey[*Chat]("dummy"),
		WithURL[*Chat](srv.URL),
		WithHTTPClient[*Chat](client),
	)
	res, err := chat.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := res.Content(), "integration reply"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestIntegrationTTSAudio(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// setup test server to mimic tts audio endpoint
	audioData := []byte("audio-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		// return raw audio bytes
		w.Write(audioData)
	}))
	defer srv.Close()

	client := srv.Client()
	tts := NewTTS(
		WithInput("hello world"),
		WithAPIKey[*TTS]("dummy"),
		WithURL[*TTS](srv.URL),
		WithHTTPClient[*TTS](client),
	)
	out, err := tts.Audio()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, audioData) {
		t.Errorf("expected audio %v, got %v", audioData, out)
	}
}

func TestIntegrationSTTTranscript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// setup test server to mimic stt transcript endpoint
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart: %v", err)
		}
		// verify fields
		if lang := r.FormValue("language"); lang == "" {
			t.Errorf("language field missing")
		}
		files := r.MultipartForm.File["file"]
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
		f, err := files[0].Open()
		if err != nil {
			t.Fatalf("failed to open uploaded file: %v", err)
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			t.Fatalf("failed to read uploaded file: %v", err)
		}
		if len(data) == 0 {
			t.Errorf("uploaded file empty")
		}

		// respond with transcript json
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"text":"integration transcript"}`))
	}))
	defer srv.Close()

	// create a temp audio file
	tmp, err := os.CreateTemp("", "stt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.WriteString("dummy audio content")
	tmp.Seek(0, io.SeekStart)

	client := srv.Client()
	stt := NewSTT(
		WithFile(tmp),
		WithFilePath(tmp.Name()),
		WithLanguage("en"),
		WithAPIKey[*STT]("dummy"),
		WithURL[*STT](srv.URL),
		WithHTTPClient[*STT](client),
	)
	res, err := stt.Transcript()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := res.Content(), "integration transcript"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
