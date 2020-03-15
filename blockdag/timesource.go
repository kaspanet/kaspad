package blockdag

import (
	"time"
)

// TimeSource is the interface to access time.
type TimeSource interface {
	// Now returns the current time.
	Now() time.Time
}

// timeSource provides an implementation of the TimeSource interface
// that simply returns the current local time.
type timeSource struct{}

// Now returns the current local time, with one second precision.
func (m *timeSource) Now() time.Time {
	return time.Unix(time.Now().Unix(), 0)
}

// NewTimeSource returns a new instance of a TimeSource
func NewTimeSource() TimeSource {
	return &timeSource{}
}
