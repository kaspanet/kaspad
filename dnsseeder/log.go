package main

import (
	"fmt"
	"github.com/btcsuite/btclog"
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

var (
	backendLog = btclog.NewBackend(logWriter{})
	LogRotator *rotator.Rotator
	logger     = backendLog.Logger("SEED")
	initiated  = false
)

func initLogRotator(logFile string) {
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

	LogRotator = r
}
