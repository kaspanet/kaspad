package grpcclient

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.RPCC)
var spawn = panics.GoroutineWrapperFunc(log)
