package database

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/kasparov/logger"

var (
	log   = logger.BackendLog.Logger("DTBS")
	spawn = panics.GoroutineWrapperFunc(log)
)
