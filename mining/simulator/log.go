package main

import (
	"log"
	"os"

	"github.com/btcsuite/btclog"
	"github.com/daglabs/btcd/rpcclient"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return len(p), nil
}

func init() {
	backendLog := btclog.NewBackend(logWriter{})
	rpclog := backendLog.Logger("RPCC")
	rpcclient.UseLogger(rpclog)

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}
