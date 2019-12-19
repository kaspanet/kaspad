package main

import (
	"github.com/kaspanet/kaspad/logs"
)

var (
	backendLog = logs.NewBackend()
	log        = backendLog.Logger("ASUB")
)
