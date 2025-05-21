package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewSlogger_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	sl := NewSlogger(buf, false, slog.LevelWarn)

	sl.Debug("debug-msg")
	if buf.Len() != 0 {
		t.Errorf("debug: want 0, got %d", buf.Len())
	}

	sl.Info("info-msg")
	if buf.Len() != 0 {
		t.Errorf("info: want 0, got %d", buf.Len())
	}

	sl.Warn("warn-msg")
	out := buf.String()
	if !strings.Contains(out, "level=WARN") {
		t.Errorf("warn: want level=WARN, got %q", out)
	}
	buf.Reset()

	sl.Error("error-msg")
	out = buf.String()
	if !strings.Contains(out, "level=ERROR") {
		t.Errorf("error: want level=ERROR, got %q", out)
	}
}

func TestSlogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	sl := NewSlogger(buf, false, slog.LevelDebug)
	ctx := sl.With("foo", "bar", "num", 42)
	ctx.Info("msg")
	out := buf.String()
	if !strings.Contains(out, "foo=bar") {
		t.Errorf("context foo: want foo=bar, got %q", out)
	}
	if !strings.Contains(out, "num=42") {
		t.Errorf("context num: want num=42, got %q", out)
	}
	if !strings.Contains(out, "msg=msg") {
		t.Errorf("msg: want msg=msg, got %q", out)
	}
}

func TestDefaultFunction_UseDefaultLogger(t *testing.T) {
	old := defaultLogger
	buf := &bytes.Buffer{}
	defaultLogger = NewSlogger(buf, false, slog.LevelError)
	defer func() { defaultLogger = old }()

	def := Default()
	def.Info("x")
	if buf.Len() != 0 {
		t.Errorf("info default: want 0, got %d", buf.Len())
	}

	def.Error("err", "k", "v")
	out := buf.String()
	if !strings.Contains(out, "level=ERROR") {
		t.Errorf("default error: want level=ERROR, got %q", out)
	}
	if !strings.Contains(out, "k=v") {
		t.Errorf("context k: want k=v, got %q", out)
	}
}

func TestNewTestLogger_InitialState(t *testing.T) {
	tl := NewTestLogger()
	// debug
	if tl.CalledDebug() {
		t.Errorf("called debug: want false, got true")
	}
	if tl.CountDebug() != 0 {
		t.Errorf("count debug: want 0, got %d", tl.CountDebug())
	}
	// info
	if tl.CalledInfo() {
		t.Errorf("called info: want false, got true")
	}
	if tl.CountInfo() != 0 {
		t.Errorf("count info: want 0, got %d", tl.CountInfo())
	}
	// warn
	if tl.CalledWarn() {
		t.Errorf("called warn: want false, got true")
	}
	if tl.CountWarn() != 0 {
		t.Errorf("count warn: want 0, got %d", tl.CountWarn())
	}
	// error
	if tl.CalledError() {
		t.Errorf("called error: want false, got true")
	}
	if tl.CountError() != 0 {
		t.Errorf("count error: want 0, got %d", tl.CountError())
	}
}

func TestTestLogger_MethodsAndCounts(t *testing.T) {
	tl := NewTestLogger()

	// debug
	tl.Debug("d")
	tl.Debug("d")
	if !tl.CalledDebug() {
		t.Errorf("called debug: want true, got false")
	}
	if tl.CountDebug() != 2 {
		t.Errorf("count debug: want 2, got %d", tl.CountDebug())
	}

	// info
	tl.Info("i")
	if !tl.CalledInfo() {
		t.Errorf("called info: want true, got false")
	}
	if tl.CountInfo() != 1 {
		t.Errorf("count info: want 1, got %d", tl.CountInfo())
	}

	// warn
	tl.Warn("w")
	tl.Warn("w")
	tl.Warn("w")
	if !tl.CalledWarn() {
		t.Errorf("called warn: want true, got false")
	}
	if tl.CountWarn() != 3 {
		t.Errorf("count warn: want 3, got %d", tl.CountWarn())
	}

	// error
	tl.Error("e")
	if !tl.CalledError() {
		t.Errorf("called error: want true, got false")
	}
	if tl.CountError() != 1 {
		t.Errorf("count error: want 1, got %d", tl.CountError())
	}
}

func TestTestLogger_With(t *testing.T) {
	tl := NewTestLogger()
	tw := tl.With("key", "value")
	if tw != tl {
		t.Errorf("with: want same logger, got different")
	}
	tl.Info("x")
	if tl.CountInfo() != 1 {
		t.Errorf("count info: want 1, got %d", tl.CountInfo())
	}
}
