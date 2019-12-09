package main

import (
	"github.com/kaspanet/kaspad/kasparov/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	log   = logger.Logger("KVSD")
	spawn = panics.GoroutineWrapperFunc(log)
)
