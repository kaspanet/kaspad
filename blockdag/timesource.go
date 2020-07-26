package blockdag

import (
	"github.com/kaspanet/kaspad/util/mstime"
)

// TimeSource is the interface to access time.
type TimeSource interface {
	// Now returns the current time.
	Now() mstime.Time
}

// timeSource provides an implementation of the TimeSource interface
// that simply returns the current local time.
type timeSource struct{}

// Now returns the current local time, with one millisecond precision.
func (m *timeSource) Now() mstime.Time {
	return mstime.Now()
}

// NewTimeSource returns a new instance of a TimeSource
func NewTimeSource() TimeSource {
	return &timeSource{}
}
