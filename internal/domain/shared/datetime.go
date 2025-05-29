package shared

import (
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
)

const (
	MDatetimeMissing string = "Missing datetime."
	MDatetimeNotPast string = "Datetime must not be in the future."
)

// Datetime wraps time with domain-specific validation for audit trails.
// Ensures temporal data integrity and prevents logical inconsistencies.
type Datetime struct{ t time.Time }

// NewDatetime creates validated datetime for historical record keeping.
// Enforces past-only timestamps for audit trail integrity.
func NewDatetime(u time.Time) (Datetime, error) {
	const op = "NewDatetime"

	d := Datetime{t: u.UTC()}
	if err := d.Validate(); err != nil {
		return Datetime{}, &kernel.Error{Operation: op, Cause: err}
	}

	return d, nil
}

// NewDatetimeNow captures current moment for timestamp generation.
// Provides consistent time source for creation and modification tracking.
func NewDatetimeNow() (Datetime, error) {
	return NewDatetime(time.Now())
}

// NewDatetimeAllowFuture enables future timestamps for scheduling features.
// Supports scheduled publishing while maintaining presence validation.
func NewDatetimeAllowFuture(u time.Time) (Datetime, error) {
	const op = "NewDatetimeAllowFuture"

	d := Datetime{t: u.UTC()}
	if err := d.validatePresent(); err != nil {
		return Datetime{}, &kernel.Error{Operation: op, Cause: err}
	}

	return d, nil
}

func (d Datetime) Time() time.Time            { return d.t }
func (d Datetime) Before(other Datetime) bool { return d.t.Before(other.t) }
func (d Datetime) After(other Datetime) bool  { return d.t.After(other.t) }
func (d Datetime) Equal(other Datetime) bool  { return d.t.Equal(other.t) }
func (d Datetime) String() string             { return d.t.Format(time.RFC3339) }

// Validate ensures datetime meets domain requirements for temporal consistency.
// Prevents future timestamps that would violate audit trail integrity.
func (d Datetime) Validate() error {
	const op = "Datetime.Validate"

	if err := d.validatePresent(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	if err := d.validateNotFuture(); err != nil {
		return &kernel.Error{Operation: op, Cause: err}
	}

	return nil
}

func (d Datetime) validatePresent() error {
	const op = "Datetime.validatePresent"

	if d.t.IsZero() {
		return &kernel.Error{Code: kernel.EInvalid, Message: MDatetimeMissing, Operation: op}
	}

	return nil
}

func (d Datetime) validateNotFuture() error {
	const op = "Datetime.validateNotFuture"

	if d.t.After(time.Now().UTC()) {
		return &kernel.Error{Code: kernel.EInvalid, Message: MDatetimeNotPast, Operation: op}
	}

	return nil
}
