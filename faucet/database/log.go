package database

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/apiserver/logger"

var (
	log   = logger.BackendLog.Logger("DTBS")
	spawn = panics.GoroutineWrapperFunc(log)
)
