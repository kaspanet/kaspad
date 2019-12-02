package database

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/apiserver/logger"

var (
	log   = logger.Logger("DTBS")
	spawn = panics.GoroutineWrapperFunc(log, logger.BackendLog)
)
