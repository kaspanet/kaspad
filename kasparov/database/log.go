package database

import "github.com/daglabs/kaspad/util/panics"
import "github.com/daglabs/kaspad/kasparov/logger"

var (
	log   = logger.Logger("DTBS")
	spawn = panics.GoroutineWrapperFunc(log)
)
