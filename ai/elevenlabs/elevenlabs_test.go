package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alnah/fla/ai"
	"github.com/alnah/fla/logger"
)

/********* Helpers *********/

type roundTripperTest func(req *http.Request) (*http.Response, error)

func (f roundTripperTest) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type marker string

const key marker = "marker"

/********* Tests *********/

func TestTTS_WithOption_Pattern(t *testing.T) {
	log := logger.New()
	ctx := context.WithValue(context.Background(), key, "value")

	hc := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		if got := req.Context().Value(key); got != "value" {
			t.Error("http request: want context inheritance from chat")
		}
		body := `{"content": [{"text": "ok", "type":"" }], "id": "1"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}

	tts := NewTTS(
		WithModel(ModelTextToSpeech),
		WithLogger(log),
		WithHTTPClient(hc),
		WithContext(ctx),
		WithAPIKey("test-api-key"),
		WithURL("http://test.com"),
		WithInput("test"),
		WithVoice(VoiceFemaleAudrey),
		WithSpeed(1.0),
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
	if tts.voice.String() != VoiceFemaleAudrey.String() {
		t.Errorf("with voice: want %q, got %q", VoiceFemaleAudrey.String(), tts.voice.String())
	}
	if tts.Speed != 1 {
		t.Errorf("with speed: want %f.2, got %f.2", 1.0, tts.Speed)
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
		WithAPIKey("test-api-key"),
		WithURL("::::"),
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
		WithContext(ctx),
		WithInput("test"),
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithAPIKey("k"),
		WithURL("http://u/"),
		WithHTTPClient(hc),
	)

	audio, err := tts.Audio()
	if err != nil {
		t.Fatalf("want no error: got %v", err)
	}
	if !bytes.Equal(audio, data) {
		t.Errorf("audio bytes: want %v, got %v", audio, data)
	}
}
