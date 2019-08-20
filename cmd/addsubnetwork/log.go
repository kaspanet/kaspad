package main

import (
	"github.com/daglabs/btcd/logs"
)

var (
	backendLog = logs.NewBackend()
	log        = backendLog.Logger("ASUB")
)
