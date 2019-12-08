package mqtt

import "github.com/kaspanet/kaspad/util/panics"
import "github.com/kaspanet/kaspad/kasparov/logger"

var (
	log   = logger.Logger("MQTT")
	spawn = panics.GoroutineWrapperFunc(log)
)
