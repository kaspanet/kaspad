package profiling

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.PROF)
var spawn = panics.GoroutineWrapperFunc(log)
var spawnAfter = panics.AfterFuncWrapperFunc(log)
