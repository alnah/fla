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
		t.Errorf("didn't want log output: %s", output)
	}
}

func TestLogger_WriteTo_Stdout(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() *Logger
	}{
		{name: "New", fn: New},
		{name: "Test", fn: Test},
	}
	for _, tc := range testCases {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		// save original stdout
		origStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		logger := tc.fn()
		logger.Info("test", "foo", "bar")

		// close the writer and restore stdout
		w.Close()
		os.Stdout = origStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "msg=test") || !strings.Contains(output, "foo=bar") {
			t.Errorf("didn't want log output: %s", output)
		}

	}
}
