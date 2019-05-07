package gowrapper

import (
	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/util/panics"
)

// Generate returns a goroutine wrapper that handles panics and write them to the log.
func Generate(log btclog.Logger) func(func()) {
	return func(f func()) {
		go func() {
			defer panics.HandlePanic(log)
			f()
		}()
	}
}
