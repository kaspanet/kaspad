package mqtt

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/kasparov/logger"

var (
	log   = logger.Logger("MQTT")
	spawn = panics.GoroutineWrapperFunc(log)
)
