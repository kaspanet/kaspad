package main

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	log   = logger.BackendLog.Logger("FAUC")
	spawn = panics.GoroutineWrapperFunc(log)
)
