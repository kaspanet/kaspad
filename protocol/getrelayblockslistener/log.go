package getrelayblockslistener

import (
	"github.com/kaspanet/kaspad/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log, _ = logger.Get(logger.SubsystemTags.GBRL)
var spawn = panics.GoroutineWrapperFunc(log)
