package mqtt

import "github.com/daglabs/btcd/util/panics"
import "github.com/daglabs/btcd/apiserver/logger"

var (
	log   = logger.BackendLog.Logger("MQTT")
	spawn = panics.GoroutineWrapperFunc(log, logger.BackendLog)
)
