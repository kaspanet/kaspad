package rpchandlers

import (
	"github.com/c4ei/kaspad/infrastructure/logger"
	"github.com/c4ei/kaspad/util/panics"
)

var log = logger.RegisterSubSystem("RPCS")
var spawn = panics.GoroutineWrapperFunc(log)
