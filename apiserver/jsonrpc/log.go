package jsonrpc

import (
	"github.com/daglabs/btcd/apiserver/logger"
	"github.com/daglabs/btcd/rpcclient"
	"github.com/daglabs/btcd/util/panics"
)

var (
	log   = logger.BackendLog.Logger("RPCC")
	spawn = panics.GoroutineWrapperFunc(log, logger.BackendLog)
)

func init() {
	rpcclient.UseLogger(log, logger.BackendLog)
}
