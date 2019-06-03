package main

import (
	"github.com/btcsuite/btclog"
	"os"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

var (
	backendLog = btclog.NewBackend(logWriter{})
	log        = backendLog.Logger("ASUB")
)
