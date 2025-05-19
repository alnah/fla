package logger

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLogger_WriteTo_Stdout(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() *slog.Logger
	}{
		{name: "New", fn: New},
		{name: "Test", fn: NewTestLogger},
	}
	for _, tc := range testCases {
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
