package storage

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Address represents a network endpoint as a host and port pair.
// It ensures that hosts and ports conform to valid syntax.
// Use ParseAddress to construct and validate instances.
type Address struct {
	Host string
	Port int
}

// ParseAddress parses a string of the form "host:port" into an Address.
// It returns an error if the format is invalid or fails validation.
func ParseAddress(s string) (Address, error) {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return Address{}, fmt.Errorf("invalid address format: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Address{}, fmt.Errorf("invalid port: %w", err)
	}
	addr := Address{Host: host, Port: port}
	if err := addr.Validate(); err != nil {
		return Address{}, err
	}
	return addr, nil
}

// String returns the Address as "host:port".
// IPv6 addresses are wrapped in brackets per standard notation.
func (a Address) String() string {
	ip := net.ParseIP(a.Host)
	if ip != nil && ip.To4() == nil {
		// IPv6 literal
		return fmt.Sprintf("[%s]:%d", a.Host, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// IsValid reports whether the Address passes all validation rules.
func (c Address) IsValid() bool {
	return c.Validate() == nil
}

// Validate enforces that Host is a non-empty valid IP or DNS name,
// and Port lies within the range 1 to 65535.
func (c Address) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is empty")
	}
	if ip := net.ParseIP(c.Host); ip == nil {
		if !isHostname(c.Host) {
			return fmt.Errorf("invalid host: %q", c.Host)
		}
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port %d out of range (1–65535)", c.Port)
	}
	return nil
}

// isHostname checks that a string follows RFC-1123 label rules:
// each label is 1–63 characters, alphanumeric or hyphen,
// and labels do not start or end with a hyphen.
func isHostname(s string) bool {
	if len(s) > 253 {
		return false
	}
	labels := strings.SplitSeq(s, ".")
	for lbl := range labels {
		if lbl == "" || len(lbl) > 63 {
			return false
		}
		if !isLetterOrDigit(lbl[0]) || !isLetterOrDigit(lbl[len(lbl)-1]) {
			return false
		}
		for i := 1; i < len(lbl)-1; i++ {
			c := lbl[i]
			if !(isLetterOrDigit(c) || c == '-') {
				return false
			}
		}
	}
	return true
}

// isLetterOrDigit returns true for ASCII letters and digits.
func isLetterOrDigit(c byte) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		('0' <= c && c <= '9')
}

// Password represents a secret credential string.
// It centralizes validation and prevents empty values.
type Password string

// ParsePassword constructs a Password after validation.
func ParsePassword(s string) (Password, error) {
	pwd := Password(s)
	if err := pwd.Validate(); err != nil {
		return "", err
	}
	return pwd, nil
}

// String returns the raw password value.
func (p Password) String() string { return string(p) }

// IsValid reports whether the Password is non-empty.
func (p Password) IsValid() bool { return p.Validate() == nil }

// Validate enforces that the Password is not empty.
func (p Password) Validate() error {
	if p == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}

// LogicalDBs represents a Redis logical database index (0–15).
type LogicalDBs int

// Int returns the underlying database index.
func (l LogicalDBs) Int() int { return int(l) }

// IsValid reports whether the index lies within 0–15.
func (l LogicalDBs) IsValid() bool { return l.Validate() == nil }

// Validate enforces that the index is between 0 and 15 inclusive.
func (l LogicalDBs) Validate() error {
	if l.Int() < 0 || l.Int() > 15 {
		return fmt.Errorf("logical db number must be between 0 and 15")
	}
	return nil
}

// Timeout represents a duration that must be positive.
type Timeout time.Duration

// ParseTimeout constructs a Timeout after validation.
func ParseTimeout(t time.Duration) (Timeout, error) {
	timeout := Timeout(t)
	if err := timeout.Validate(); err != nil {
		return 0, err
	}
	return timeout, nil
}

// Duration returns the raw time.Duration value.
func (t Timeout) Duration() time.Duration {
	return time.Duration(t)
}

// IsValid reports whether the Timeout is positive.
func (t Timeout) IsValid() bool { return t.Validate() == nil }

// Validate enforces that the Timeout value is greater than zero.
func (t Timeout) Validate() error {
	if t <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}
