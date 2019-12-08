package mqtt

import "github.com/daglabs/kaspad/util/panics"
import "github.com/daglabs/kaspad/kasparov/logger"

var (
	log   = logger.Logger("MQTT")
	spawn = panics.GoroutineWrapperFunc(log)
)
