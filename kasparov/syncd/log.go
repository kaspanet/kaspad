package main

import (
	"github.com/daglabs/btcd/kasparov/logger"
	"github.com/daglabs/btcd/util/panics"
)

var (
	log   = logger.Logger("KVSD")
	spawn = panics.GoroutineWrapperFunc(log)
)
