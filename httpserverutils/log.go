package httpserverutils

import "github.com/daglabs/kaspad/util/panics"
import "github.com/daglabs/kaspad/kasparov/logger"

var (
	log   = logger.BackendLog.Logger("UTIL")
	spawn = panics.GoroutineWrapperFunc(log)
)
