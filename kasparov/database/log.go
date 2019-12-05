package database

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/kasparov/logger"

var (
	log   = logger.Logger("DTBS")
	spawn = panics.GoroutineWrapperFunc(log)
)
