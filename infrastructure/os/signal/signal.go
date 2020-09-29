// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package signal

import (
	"os"
	"os/signal"
)

// ShutdownRequestChannel is used to initiate shutdown from one of the
// subsystems using the same code paths as when an interrupt signal is received.
var ShutdownRequestChannel = make(chan struct{})

// interruptSignals defines the default signals to catch in order to do a proper
// shutdown. This may be modified during init depending on the platform.
var interruptSignals = []os.Signal{os.Interrupt}

// InterruptListener listens for OS Signals such as SIGINT (Ctrl+C) and shutdown
// requests from shutdownRequestChannel. It returns a channel that is closed
// when either signal is received.
func InterruptListener() chan struct{} {
	c := make(chan struct{})
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, interruptSignals...)

		// Listen for initial shutdown signal and close the returned
		// channel to notify the caller.
		select {
		case sig := <-interruptChannel:
			kasdLog.Infof("Received signal (%s). Shutting down...",
				sig)

		case <-ShutdownRequestChannel:
			kasdLog.Info("Shutdown requested. Shutting down...")
		}
		close(c)

		// Listen for repeated signals and display a message so the user
		// knows the shutdown is in progress and the process is not
		// hung.
		for {
			select {
			case sig := <-interruptChannel:
				kasdLog.Infof("Received signal (%s). Already "+
					"shutting down...", sig)

			case <-ShutdownRequestChannel:
				kasdLog.Info("Shutdown requested. Already " +
					"shutting down...")
			}
		}
	}()

	return c
}

// InterruptRequested returns true when the channel returned by
// InterruptListener was closed. This simplifies early shutdown slightly since
// the caller can just use an if statement instead of a select.
func InterruptRequested(interrupted <-chan struct{}) bool {
	select {
	case <-interrupted:
		return true
	default:
	}

	return false
}
