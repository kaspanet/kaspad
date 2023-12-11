package rpc

import (
	"github.com/fabbez/topiad/infrastructure/logger"
	"github.com/fabbez/topiad/util/panics"
)

var log = logger.RegisterSubSystem("RPCS")
var spawn = panics.GoroutineWrapperFunc(log)
