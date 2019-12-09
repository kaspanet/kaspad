package main

import (
	"github.com/kaspanet/kaspad/kasparov/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	log   = logger.Logger("KVSV")
	spawn = panics.GoroutineWrapperFunc(log)
)
