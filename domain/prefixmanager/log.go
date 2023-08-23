package prefixmanager

import (
	"github.com/c4ei/kaspad/infrastructure/logger"
	"github.com/c4ei/kaspad/util/panics"
)

var log = logger.RegisterSubSystem("PRFX")
var spawn = panics.GoroutineWrapperFunc(log)
