package main

import (
	"fmt"
	"github.com/daglabs/btcd/logs"
	"github.com/daglabs/btcd/util/panics"
	"github.com/jrick/logrotate/rotator"
	"os"
	"path/filepath"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	if initiated {
		os.Stdout.Write(p)
		LogRotator.Write(p)
	}
	return len(p), nil
}

// errLogWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type errLogWriter struct{}

func (errLogWriter) Write(p []byte) (n int, err error) {
	if initiated {
		os.Stdout.Write(p)
		ErrLogRotator.Write(p)
	}
	return len(p), nil
}

var (
	backendLog = logs.NewBackend([]*logs.BackendWriter{
		logs.NewAllLevelsBackendWriter(logWriter{}),
		logs.NewErrorBackendWriter(errLogWriter{}),
	})
	ErrLogRotator, LogRotator *rotator.Rotator
	log        = backendLog.Logger("TXGN")
	spawn      = panics.GoroutineWrapperFunc(log)
	initiated  = false
)

func initLogRotator(logFile string) *rotator.Rotator {
	initiated = true
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %s\n", err)
		os.Exit(1)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %s\n", err)
		os.Exit(1)
	}
	return r
}

func initLogRotators(logFile, errLogFile string) {
	LogRotator = initLogRotator(logFile)
	ErrLogRotator = initLogRotator(errLogFile)
}
