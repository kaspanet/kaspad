package server

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/kasparov/logger"

var (
	log   = logger.Logger("REST")
	spawn = panics.GoroutineWrapperFunc(log)
)
