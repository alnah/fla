package anthropic

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
	"github.com/alnah/fla/clog"
)

// roundTripperTest allows injecting custom RoundTrip behavior.
type roundTripperTest func(req *http.Request) (*http.Response, error)

func (f roundTripperTest) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type marker string

const key marker = "marker"

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
		WithModel(ModelReasoning),
		WithLogger(log),
		WithHTTPClient(hc),
		WithContext(ctx),
		WithAPIKey("test-api-key"),
		WithURL("http://test.com"),
		WithMessages(msg),
		WithMaxTokens(1234),
		WithSystem("system-msg"),
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
	if chat.System != "system-msg" {
		t.Errorf("with system: want %q, got %q", "systemmsg", chat.System)
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

	if chat.System == "got change" {
		t.Error("chat system: want no change, got change")
	}
}

func TestChat_SetSystemAndAddMessage_FluentPattern(t *testing.T) {
	chat := NewChat().
		SetSystem("system-msg").
		AddMessage(ai.RoleUser, "user-msg").
		AddMessage(ai.RoleAssistant, "assistant-msg")

	if chat.System != "system-msg" {
		t.Errorf("chat system: want \"system-msg\", got %q", chat.System)
	}
	if got := len(chat.Messages); got != 2 {
		t.Errorf("chat messages length: want %d, got %d", 2, got)
	}
	got1, got2, want1, want2 := chat.Messages[0].Content, chat.Messages[1].Content, "user-msg", "assistant-msg"
	if got1 != want1 || got2 != want2 {
		t.Errorf("messages: want [%s, %s], got [%s, %s]", want1, want2, got1, got2)
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
		WithAPIKey("test"),
		WithURL("::::"),
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
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithContext(ctx),
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "test"}}),
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
		WithAPIKey("test-api-key"),
		WithURL("https://test.com"),
		WithHTTPClient(hc),
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
