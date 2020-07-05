package blockrelay

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.BLKR)
var spawn = panics.GoroutineWrapperFunc(log)
