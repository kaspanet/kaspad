package server

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/apiserver/logger"

var (
	log   = logger.Logger("REST")
	spawn = panics.GoroutineWrapperFunc(log)
)
