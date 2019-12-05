package main

import (
	"github.com/daglabs/btcd/kasparov/logger"
	"github.com/daglabs/btcd/util/panics"
)

var (
	log   = logger.Logger("APIS")
	spawn = panics.GoroutineWrapperFunc(log)
)
