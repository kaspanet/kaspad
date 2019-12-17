package httpserverutils

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	log   = logger.BackendLog.Logger("UTIL")
	spawn = panics.GoroutineWrapperFunc(log)
)
