package main

import (
	"github.com/daglabs/kaspad/logger"
	"github.com/daglabs/kaspad/util/panics"
)

var (
	log   = logger.BackendLog.Logger("FAUC")
	spawn = panics.GoroutineWrapperFunc(log)
)
