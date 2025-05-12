package logger

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLogger_Debug_Output(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})

	logger := NewWithHandler(handler)
	logger.Debug("testing", "key", "value")
	output := buf.String()

	if !strings.Contains(output, "msg=testing") || !strings.Contains(output, "key=value") {
		t.Errorf("unexpected log output: %s", output)
	}
}

func TestLoggerNew_WritesTo_Stdout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// save original stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := New()
	logger.Info("hello", "foo", "bar")

	// close the writer and restore stdout
	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "msg=hello") || !strings.Contains(output, "foo=bar") {
		t.Errorf("unexpected log output: %s", output)
	}
}
