package logger

import "sync"

// testLogger records method invocations and counts for use in tests,
// ensuring code under test logs expected events without real output.
type testLogger struct {
	mu       sync.Mutex
	debugLog struct {
		called bool
		count  int
	}
	infoLog struct {
		called bool
		count  int
	}
	warnLog struct {
		called bool
		count  int
	}
	errorLog struct {
		called bool
		count  int
	}
}

// NewTestLogger returns a Logger that tracks calls and counts,
// enabling assertions on whether and how often each level was used.
func NewTestLogger() *testLogger {
	return &testLogger{}
}

// Debug marks that a debug-level log was emitted.
func (t *testLogger) Debug(msg string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.debugLog.called = true
	t.debugLog.count++
}

// Info marks that an info-level log was emitted.
func (t *testLogger) Info(msg string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.infoLog.called = true
	t.infoLog.count++
}

// Warn marks that a warn-level log was emitted.
func (t *testLogger) Warn(msg string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.warnLog.called = true
	t.warnLog.count++
}

// Error marks that an error-level log was emitted.
func (t *testLogger) Error(msg string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorLog.called = true
	t.errorLog.count++
}

// With returns the same testLogger, ignoring additional context fields
// since tests focus solely on invocation counts.
func (t *testLogger) With(args ...any) Logger {
	return t
}

// CalledDebug reports whether Debug was ever called.
func (t *testLogger) CalledDebug() bool { return t.debugLog.called }

// CountDebug returns how many times Debug was called.
func (t *testLogger) CountDebug() int { return t.debugLog.count }

// CalledInfo reports whether Info was ever called.
func (t *testLogger) CalledInfo() bool { return t.infoLog.called }

// CountInfo returns how many times Info was called.
func (t *testLogger) CountInfo() int { return t.infoLog.count }

// CalledWarn reports whether Warn was ever called.
func (t *testLogger) CalledWarn() bool { return t.warnLog.called }

// CountWarn returns how many times Warn was called.
func (t *testLogger) CountWarn() int { return t.warnLog.count }

// CalledError reports whether Error was ever called.
func (t *testLogger) CalledError() bool { return t.errorLog.called }

// CountError returns how many times Error was called.
func (t *testLogger) CountError() int { return t.errorLog.count }
