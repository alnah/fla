package locale

import "fmt"

// Lang is the set of supported 2-letter language codes.
type Lang string

const (
	FR Lang = "fr"
	PT Lang = "pt"
	EN Lang = "en"
)

func (l Lang) String() string { return string(l) }

func (l Lang) Validate() error {
	switch l {
	case FR, PT, EN:
		return nil
	}
	return fmt.Errorf("invalid language: %q", l)
}

// Set implements flag.Value so it can be bind directly in config.BindFlags().
func (l *Lang) Set(s string) error {
	*l = Lang(s)
	return l.Validate()
}

// UnmarshalText implements encoding.TextUnmarshaler so JSON decode will validate automatically
func (l *Lang) UnmarshalText(text []byte) error {
	*l = Lang(text)
	return l.Validate()
}
