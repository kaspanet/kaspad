package httpserverutils

import (
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kasparov/logger"
)

var (
	log   = logger.BackendLog.Logger("UTIL")
	spawn = panics.GoroutineWrapperFunc(log)
)
