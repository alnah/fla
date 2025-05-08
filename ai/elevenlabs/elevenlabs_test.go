package elevenlabs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

type key string

const marker key = "marker"

/********* Tests *********/

func TestNewTTS_Defaults(t *testing.T) {
	// ensure env API key is picked up
	os.Setenv(apiKeyFromEnv, "env-key")
	defer os.Unsetenv(apiKeyFromEnv)

	tts := NewTTS()
	if tts.Model != ModelTextToSpeech {
		t.Errorf("default Model = %q; want %q", tts.Model, ModelTextToSpeech)
	}
	if tts.voice != VoiceMaleNicolas {
		t.Errorf("default voice = %q; want %q", tts.voice, VoiceMaleNicolas)
	}
	if tts.apiKey != "env-key" {
		t.Errorf("default apiKey = %q; want %q", tts.apiKey, "env-key")
	}
	if tts.method != http.MethodPost {
		t.Errorf("default method = %q; want POST", tts.method)
	}
	wantURL := gateway + pathTTS
	if tts.url != wantURL {
		t.Errorf("default url = %q; want %q", tts.url, wantURL)
	}
}

func TestTTS_OptionSetters(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), marker, "value")
	client := &http.Client{Timeout: time.Second}
	customModel := model("mymodel")
	customVoice := VoiceMaleGuillaume
	tts := NewTTS(
		WithModel(customModel),
		WithLogger(log),
		WithHTTPClient(client),
		WithContext(ctx),
		WithAPIKey("key123"),
		WithURL("http://custom/"),
		WithInput("hello"),
		WithVoice(customVoice),
		WithSpeed(1.1),
	)

	if tts.Model != customModel {
		t.Errorf("WithModel: got %q, want %q", tts.Model, customModel)
	}
	if tts.log != log {
		t.Error("WithLogger did not set logger")
	}
	if tts.hc != client {
		t.Error("WithHTTPClient did not set http.Client")
	}
	if tts.ctx != ctx {
		t.Error("WithContext did not set context")
	}
	if tts.apiKey != "key123" {
		t.Errorf("WithAPIKey: got %q, want key123", tts.apiKey)
	}
	if tts.url != "http://custom/" {
		t.Errorf("WithURL: got %q, want http://custom/", tts.url)
	}
	if tts.Input != "hello" {
		t.Errorf("WithInput: got %q, want hello", tts.Input)
	}
	if tts.voice != customVoice {
		t.Errorf("WithVoice: got %q, want %q", tts.voice, customVoice)
	}
	if tts.Speed != 1.1 {
		t.Errorf("WithSpeed: got %v, want 1.1", tts.Speed)
	}
}

func TestAudio_NoInput(t *testing.T) {
	tts := NewTTS()
	_, err := tts.Audio()
	if err == nil {
		t.Errorf("expected an error, but got nil")
	}
}

func TestAudio_HTTPClientError(t *testing.T) {
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network fail")
	})}
	tts := NewTTS(
		WithInput("hi"),
		WithAPIKey("k"),
		WithURL("http://u/"),
		WithHTTPClient(client),
	)
	_, err := tts.Audio()
	if err == nil {
		t.Error("expected an error, but got nil")
	}
}

func TestAudio_HTTPStatusError(t *testing.T) {
	// simulate ElevenLabs API returning HTTP 429
	body := `{"detail":{"status":"rate_limit","message":"too many requests"}}`
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	tts := NewTTS(
		WithInput("oops"),
		WithAPIKey("k"),
		WithURL("http://u/"),
		WithHTTPClient(client),
	)

	_, err := tts.Audio()
	if err == nil {
		t.Fatal("expected error for HTTP 429")
	}
	ele, ok := err.(*ai.AIError)
	if !ok {
		t.Fatalf("expected *ai.AIError; got %T", err)
	}
	errMsg := ele.Error()
	if !strings.Contains(errMsg, "429") {
		t.Errorf("error missing status code, got %q", errMsg)
	}
}

func TestAudio_Non200ProducesAudioError(t *testing.T) {
	// transport that always returns 500
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(`{"detail":{"status":"x","message":"y"}}`)),
		}, nil
	})}

	tts := NewTTS(
		WithInput("foo"),
		WithAPIKey("key"),
		WithURL("http://example/"),
		WithHTTPClient(client),
	)

	_, err := tts.Audio()
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}

	var e *ai.AIError
	if !errors.As(err, &e) {
		t.Fatalf("expected *ai.AIError, got %T", err)
	}
}

func TestAudio_Success(t *testing.T) {
	data := []byte("audio-bytes")
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		// ensure headers were added by transport.Chain
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type header = %q; want application/json", req.Header.Get("Content-Type"))
		}
		if req.Header.Get("xi-api-key") == "" {
			t.Error("xi-api-key header missing")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})}
	tts := NewTTS(
		WithInput("hey"),
		WithAPIKey("k"),
		WithURL("http://u/"),
		WithHTTPClient(client),
	)
	out, err := tts.Audio()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Errorf("Audio bytes = %v; want %v", out, data)
	}
}

func TestIntegrationAudio(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	audioData := []byte("live-audio")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s; want POST", r.Method)
		}
		w.Write(audioData)
	}))
	defer srv.Close()

	tts := NewTTS(
		WithInput("integration"),
		WithAPIKey("dummy"),
		WithURL(srv.URL+"/"),
		WithHTTPClient(srv.Client()),
	)
	out, err := tts.Audio()
	if err != nil {
		t.Fatalf("Audio integration failed: %v", err)
	}
	if !bytes.Equal(out, audioData) {
		t.Errorf("integration Audio = %v; want %v", out, audioData)
	}
}
