package server

import "github.com/kaspanet/kaspad/util/panics"
import "github.com/kaspanet/kaspad/kasparov/logger"

var (
	log   = logger.Logger("REST")
	spawn = panics.GoroutineWrapperFunc(log)
)
