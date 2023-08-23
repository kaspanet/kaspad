package prefixmanager

import (
	"github.com/c4ei/yunseokyeol/infrastructure/logger"
	"github.com/c4ei/yunseokyeol/util/panics"
)

var log = logger.RegisterSubSystem("PRFX")
var spawn = panics.GoroutineWrapperFunc(log)
