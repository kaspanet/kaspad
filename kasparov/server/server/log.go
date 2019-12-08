package server

import "github.com/daglabs/kaspad/util/panics"
import "github.com/daglabs/kaspad/kasparov/logger"

var (
	log   = logger.Logger("REST")
	spawn = panics.GoroutineWrapperFunc(log)
)
