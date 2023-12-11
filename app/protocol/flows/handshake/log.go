package handshake

import (
	"github.com/fabbez/topiad/infrastructure/logger"
	"github.com/fabbez/topiad/util/panics"
)

var log = logger.RegisterSubSystem("PROT")
var spawn = panics.GoroutineWrapperFunc(log)
