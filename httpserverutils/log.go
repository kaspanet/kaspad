package httpserverutils

import "github.com/kaspanet/kaspad/util/panics"
import "github.com/kaspanet/kaspad/kasparov/logger"

var (
	log   = logger.BackendLog.Logger("UTIL")
	spawn = panics.GoroutineWrapperFunc(log)
)
