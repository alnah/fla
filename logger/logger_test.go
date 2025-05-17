package logger

import (
	"bytes"
	"log/slog"
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
