package connmanager

import (
	"github.com/c4ei/YunSeokYeol/infrastructure/logger"
	"github.com/c4ei/YunSeokYeol/util/panics"
)

var log = logger.RegisterSubSystem("CMGR")
var spawn = panics.GoroutineWrapperFunc(log)
