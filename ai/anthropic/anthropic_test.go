package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

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

func TestChatGenericOptions(t *testing.T) {
	log := clog.New()
	ctx := context.WithValue(context.Background(), key, "value")
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		// verify context propagation
		if req.Context().Value(key) != "value" {
			t.Error("HTTP request did not inherit context from Chat")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"content":[{"text":"ok","type":""}],"id":"1"}`)),
		}, nil
	})}
	msgs := []ai.Message{{Role: ai.RoleUser, Content: "hello"}}

	chat := NewChat(
		WithModel(ModelReasoning),
		WithLogger(log),
		WithHTTPClient(client),
		WithContext(ctx),
		WithAPIKey("testkey"),
		WithURL("http://example.com"),
		WithMessages(msgs),
		WithMaxTokens(1234),
		WithSystem("systemmsg"),
		WithTemperature(0.5),
	)

	if chat.Model != ModelReasoning {
		t.Errorf("WithModel: got %v, want %v", chat.Model, ModelReasoning)
	}
	if chat.log != log {
		t.Error("WithLogger did not set logger")
	}
	if chat.hc != client {
		t.Error("WithHTTPClient did not set HTTP client")
	}
	if chat.ctx != ctx {
		t.Error("WithContext did not set context")
	}
	if chat.apiKey != "testkey" {
		t.Errorf("WithAPIKey: got %q, want testkey", chat.apiKey)
	}
	if chat.url != "http://example.com" {
		t.Errorf("WithURL: got %q, want http://example.com", chat.url)
	}
	if chat.MaxTokens != 1234 {
		t.Errorf("WithMaxTokens: got %d, want 1234", chat.MaxTokens)
	}
	if chat.System != "systemmsg" {
		t.Errorf("WithSystem: got %q, want systemmsg", chat.System)
	}
	if chat.Temperature != 0.5 {
		t.Errorf("WithTemperature: got %v, want 0.5", chat.Temperature)
	}
	if len(chat.Messages) != 1 || chat.Messages[0] != msgs[0] {
		t.Errorf("WithMessages: got %+v, want %+v", chat.Messages, msgs)
	}
}

func TestSetSystemOnce(t *testing.T) {
	chat := NewChat()
	chat.SetSystem("should not change").SetSystem("has changed")

	if chat.System == "has changed" {
		t.Errorf("system instructions should not change, but got %s", chat.System)
	}
}

func TestAddMessage_Fluent(t *testing.T) {
	chat := NewChat()
	chat.AddMessage(ai.RoleUser, "one").AddMessage(ai.RoleAssistant, "two")
	if len(chat.Messages) != 2 {
		t.Fatalf("AddMessage appended %d messages; want 2", len(chat.Messages))
	}
	if chat.Messages[0].Content != "one" || chat.Messages[1].Content != "two" {
		t.Errorf("Messages = %+v; want [one, two]", chat.Messages)
	}
}

func TestCompletion_NoMessages(t *testing.T) {
	chat := NewChat()
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestCompletion_HTTPClientError(t *testing.T) {
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network fail")
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "hi"}}),
		WithAPIKey("k"),
		WithURL("u"),
		WithHTTPClient(client),
	)
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected error for HTTP client failure")
	}
}

func TestCompletion_HTTPErrorStatus(t *testing.T) {
	var httpErr ai.HTTPError
	httpErr.Message = "msg"
	httpErr.Type = "t"

	b, _ := json.Marshal(httpErr)
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(bytes.NewReader(b))}, nil
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "hi"}}),
		WithAPIKey("k"),
		WithURL("u"),
		WithHTTPClient(client),
	)
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected HTTPError for non-200 status")
	}
	_, ok := err.(*ai.AIError)
	if !ok {
		t.Fatalf("expected *ai.AIError, got %T", err)
	}
}

func TestCompletion_InvalidJSON(t *testing.T) {
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("invalid"))}, nil
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "x"}}),
		WithAPIKey("k"),
		WithURL("u"),
		WithHTTPClient(client),
	)
	_, err := chat.Completion()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCompletion_Success(t *testing.T) {
	resp := `{"content":[{"text":"hello","type":""}],"id":"id123"}`
	client := &http.Client{Transport: roundTripperTest(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(resp))}, nil
	})}
	chat := NewChat(
		WithMessages([]ai.Message{{Role: ai.RoleUser, Content: "x"}}),
		WithAPIKey("k"),
		WithURL("u"),
		WithHTTPClient(client),
	)
	out, err := chat.Completion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out.Content(); got != "hello" {
		t.Errorf("Content = %q; want hello", got)
	}
}

func TestCompletion_ContentEmpty(t *testing.T) {
	var c ai.Completion
	if c.Content() != "" {
		t.Errorf("expected empty content, got %q", c.Content())
	}
}
