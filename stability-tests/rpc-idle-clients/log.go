package main

import (
	"github.com/c4ei/kaspad/infrastructure/logger"
	"github.com/c4ei/kaspad/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("RPIC")
	spawn      = panics.GoroutineWrapperFunc(log)
)
