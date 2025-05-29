package kernel

import "time"

// Clock exposes the current UTC time.
// Domain code depends only on this interface.
type Clock interface {
	Now() time.Time
}
