package main

import (
	"log"

	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/rpcclient"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	log.Print(string(p))
	return len(p), nil
}

func enableRPCLogging() {
	backendLog := btclog.NewBackend(logWriter{})
	rpclog := backendLog.Logger("RPCC")
	rpclog.SetLevel(btclog.LevelTrace)
	rpcclient.UseLogger(rpclog)

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}
