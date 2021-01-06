package testtools

import (
	"time"
)

// WaitTillAllCompleteOrTimeout waits until all the provided channels has been written to,
// or until a timeout period has passed.
// Returns true iff all channels returned in the allotted time.
func WaitTillAllCompleteOrTimeout(timeoutDuration time.Duration, chans ...chan struct{}) (ok bool) {
	timeout := time.After(timeoutDuration)

	for _, c := range chans {
		select {
		case <-c:
			continue
		case <-timeout:
			return false
		}
	}

	return true
}
