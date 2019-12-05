package main

import (
	"github.com/daglabs/btcd/apiserver/logger"
	"github.com/daglabs/btcd/util/panics"
)

var (
	log   = logger.Logger("SYNC")
	spawn = panics.GoroutineWrapperFunc(log)
)
