package httpserverutils

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/kasparov/logger"

var (
	log   = logger.BackendLog.Logger("UTIL")
	spawn = panics.GoroutineWrapperFunc(log)
)
