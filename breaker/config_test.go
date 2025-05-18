package breaker

import (
	"testing"
	"time"
)

func TestThirdPartyConfig(t *testing.T) {
	cfg := ThirdPartyConfig()
	want := Config{
		FailureThreshold: 3,
		FailureWindow:    0,
		SuccessThreshold: 1,
		OpenTimeout:      60 * time.Second,
		MaxOpenTimeout:   5 * time.Minute,
		BackoffMaxExp:    4,
	}

	if cfg.FailureThreshold != want.FailureThreshold {
		t.Errorf("failure threshold: want %d, got %d",
			want.FailureThreshold, cfg.FailureThreshold)
	}
	if cfg.FailureWindow != want.FailureWindow {
		t.Errorf("failure window: want %v, got %v",
			want.FailureWindow, cfg.FailureWindow)
	}
	if cfg.SuccessThreshold != want.SuccessThreshold {
		t.Errorf("success threshold: want %d, got %d",
			want.SuccessThreshold, cfg.SuccessThreshold)
	}
	if cfg.OpenTimeout != want.OpenTimeout {
		t.Errorf("open timeout: want %v, got %v",
			want.OpenTimeout, cfg.OpenTimeout)
	}
	if cfg.MaxOpenTimeout != want.MaxOpenTimeout {
		t.Errorf("max open timeout: want %v, got %v",
			want.MaxOpenTimeout, cfg.MaxOpenTimeout)
	}
	if cfg.BackoffMaxExp != want.BackoffMaxExp {
		t.Errorf("backoff max exp: want %d, got %d",
			want.BackoffMaxExp, cfg.BackoffMaxExp)
	}
	if cfg.Clock == nil {
		t.Error("clock: want non-nil, got nil")
	}
}

func TestWebAPIConfig(t *testing.T) {
	cfg := WebAPIConfig()
	want := Config{
		FailureThreshold: 5,
		FailureWindow:    10 * time.Second,
		SuccessThreshold: 2,
		OpenTimeout:      20 * time.Second,
		MaxOpenTimeout:   60 * time.Second,
		BackoffMaxExp:    2,
	}

	if cfg.FailureThreshold != want.FailureThreshold {
		t.Errorf("failure threshold: want %d, got %d",
			want.FailureThreshold, cfg.FailureThreshold)
	}
	if cfg.FailureWindow != want.FailureWindow {
		t.Errorf("failure window: want %v, got %v",
			want.FailureWindow, cfg.FailureWindow)
	}
	if cfg.SuccessThreshold != want.SuccessThreshold {
		t.Errorf("success threshold: want %d, got %d",
			want.SuccessThreshold, cfg.SuccessThreshold)
	}
	if cfg.OpenTimeout != want.OpenTimeout {
		t.Errorf("open timeout: want %v, got %v",
			want.OpenTimeout, cfg.OpenTimeout)
	}
	if cfg.MaxOpenTimeout != want.MaxOpenTimeout {
		t.Errorf("max open timeout: want %v, got %v",
			want.MaxOpenTimeout, cfg.MaxOpenTimeout)
	}
	if cfg.BackoffMaxExp != want.BackoffMaxExp {
		t.Errorf("backoff max exp: want %d, got %d",
			want.BackoffMaxExp, cfg.BackoffMaxExp)
	}
	if cfg.Clock == nil {
		t.Error("clock: want non-nil, got nil")
	}
}

func TestLowQPSConfig(t *testing.T) {
	cfg := LowQPSConfig()
	want := Config{
		FailureThreshold: 5,
		FailureWindow:    0,
		SuccessThreshold: 1,
		OpenTimeout:      30 * time.Second,
		BackoffMaxExp:    0,
	}

	if cfg.FailureThreshold != want.FailureThreshold {
		t.Errorf("failure threshold: want %d, got %d",
			want.FailureThreshold, cfg.FailureThreshold)
	}
	if cfg.FailureWindow != want.FailureWindow {
		t.Errorf("failure window: want %v, got %v",
			want.FailureWindow, cfg.FailureWindow)
	}
	if cfg.SuccessThreshold != want.SuccessThreshold {
		t.Errorf("success threshold: want %d, got %d",
			want.SuccessThreshold, cfg.SuccessThreshold)
	}
	if cfg.OpenTimeout != want.OpenTimeout {
		t.Errorf("open timeout: want %v, got %v",
			want.OpenTimeout, cfg.OpenTimeout)
	}
	if cfg.BackoffMaxExp != want.BackoffMaxExp {
		t.Errorf("backoff max exp: want %d, got %d",
			want.BackoffMaxExp, cfg.BackoffMaxExp)
	}
	if cfg.Clock == nil {
		t.Error("clock: want non-nil, got nil")
	}
}

func TestHighQPSConfig(t *testing.T) {
	cfg := HighQPSConfig()
	want := Config{
		FailureThreshold: 20,
		FailureWindow:    5 * time.Second,
		SuccessThreshold: 3,
		OpenTimeout:      10 * time.Second,
		MaxOpenTimeout:   90 * time.Second,
		BackoffMaxExp:    3,
	}

	if cfg.FailureThreshold != want.FailureThreshold {
		t.Errorf("failure threshold: want %d, got %d",
			want.FailureThreshold, cfg.FailureThreshold)
	}
	if cfg.FailureWindow != want.FailureWindow {
		t.Errorf("failure window: want %v, got %v",
			want.FailureWindow, cfg.FailureWindow)
	}
	if cfg.SuccessThreshold != want.SuccessThreshold {
		t.Errorf("success threshold: want %d, got %d",
			want.SuccessThreshold, cfg.SuccessThreshold)
	}
	if cfg.OpenTimeout != want.OpenTimeout {
		t.Errorf("open timeout: want %v, got %v",
			want.OpenTimeout, cfg.OpenTimeout)
	}
	if cfg.MaxOpenTimeout != want.MaxOpenTimeout {
		t.Errorf("max open timeout: want %v, got %v",
			want.MaxOpenTimeout, cfg.MaxOpenTimeout)
	}
	if cfg.BackoffMaxExp != want.BackoffMaxExp {
		t.Errorf("backoff max exp: want %d, got %d",
			want.BackoffMaxExp, cfg.BackoffMaxExp)
	}
	if cfg.Clock == nil {
		t.Error("clock: want non-nil, got nil")
	}
}

func TestCriticalPathConfig(t *testing.T) {
	cfg := CriticalPathConfig()
	want := Config{
		FailureThreshold: 10,
		FailureWindow:    5 * time.Second,
		SuccessThreshold: 2,
		OpenTimeout:      5 * time.Second,
		MaxOpenTimeout:   20 * time.Second,
		BackoffMaxExp:    1,
	}

	if cfg.FailureThreshold != want.FailureThreshold {
		t.Errorf("failure threshold: want %d, got %d",
			want.FailureThreshold, cfg.FailureThreshold)
	}
	if cfg.FailureWindow != want.FailureWindow {
		t.Errorf("failure window: want %v, got %v",
			want.FailureWindow, cfg.FailureWindow)
	}
	if cfg.SuccessThreshold != want.SuccessThreshold {
		t.Errorf("success threshold: want %d, got %d",
			want.SuccessThreshold, cfg.SuccessThreshold)
	}
	if cfg.OpenTimeout != want.OpenTimeout {
		t.Errorf("open timeout: want %v, got %v",
			want.OpenTimeout, cfg.OpenTimeout)
	}
	if cfg.MaxOpenTimeout != want.MaxOpenTimeout {
		t.Errorf("max open timeout: want %v, got %v",
			want.MaxOpenTimeout, cfg.MaxOpenTimeout)
	}
	if cfg.BackoffMaxExp != want.BackoffMaxExp {
		t.Errorf("backoff max exp: want %d, got %d",
			want.BackoffMaxExp, cfg.BackoffMaxExp)
	}
	if cfg.Clock == nil {
		t.Error("clock: want non-nil, got nil")
	}
}

func TestConfigZeroDefaults(t *testing.T) {
	defaultCfg := Config{}
	br := New(defaultCfg)
	got := br.(*breaker).cfg

	if got.FailureThreshold <= 0 {
		t.Errorf("failure threshold default: want > 0, got %d",
			got.FailureThreshold)
	}
	if got.SuccessThreshold <= 0 {
		t.Errorf("success threshold default: want > 0, got %d",
			got.SuccessThreshold)
	}
	if got.OpenTimeout <= 0 {
		t.Errorf("open timeout default: want > 0, got %v",
			got.OpenTimeout)
	}
	if got.Clock == nil {
		t.Error("clock default: want non-nil, got nil")
	}
}
