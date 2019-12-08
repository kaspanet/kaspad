package main

import (
	"github.com/daglabs/kaspad/logs"
)

var (
	backendLog = logs.NewBackend()
	log        = backendLog.Logger("ASUB")
)
