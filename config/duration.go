package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a thin wrapper around time.Duration that
// unmarshals from and marshals to a string like "30s", "5m".
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration should be a string, got %s: %w", string(b), err)
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(dur)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	s := time.Duration(d).String()
	return json.Marshal(s)
}

// ToTimeDuration casts Duration back to time.Duration.
func (d Duration) ToTimeDuration() time.Duration {
	return time.Duration(d)
}
