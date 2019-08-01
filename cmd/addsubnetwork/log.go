package main

import (
	"github.com/daglabs/btcd/logs"
	"os"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

var (
	backendLog = logs.NewBackend([]*logs.BackendWriter{
		logs.NewAllLevelsBackendWriter(logWriter{}),
	})
	log        = backendLog.Logger("ASUB")
)
