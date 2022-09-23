package transactionrelay

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/panics"
)

var log = logger.RegisterSubSystem("TXRL")
var spawn = panics.GoroutineWrapperFunc(log)
