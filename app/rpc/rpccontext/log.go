package rpccontext

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.RPCS)
var spawn = panics.GoroutineWrapperFunc(log)
