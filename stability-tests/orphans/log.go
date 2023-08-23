package main

import (
	"github.com/c4ei/kaspad/infrastructure/logger"
	"github.com/c4ei/kaspad/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("ORPH")
	spawn      = panics.GoroutineWrapperFunc(log)
)
