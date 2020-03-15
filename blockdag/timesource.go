// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

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

// Ensure the timeSource type implements the TimeSource interface.
var _ TimeSource = (*timeSource)(nil)

// Now returns the current local time, with one second precision.
func (m *timeSource) Now() time.Time {
	return time.Unix(time.Now().Unix(), 0)
}

// NewTimeSource returns a new instance of a TimeSource
func NewTimeSource() TimeSource {
	return &timeSource{}
}
