package flowcontext

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.PROT)
var spawn = panics.GoroutineWrapperFunc(log)
