package main

import (
	"github.com/c4ei/yunseokyeol/infrastructure/logger"
	"github.com/c4ei/yunseokyeol/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("IFLG")
	spawn      = panics.GoroutineWrapperFunc(log)
)
