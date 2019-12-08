package jsonrpc

import (
	"github.com/daglabs/kaspad/kasparov/logger"
	"github.com/daglabs/kaspad/rpcclient"
	"github.com/daglabs/kaspad/util/panics"
)

var (
	log   = logger.BackendLog.Logger("RPCC")
	spawn = panics.GoroutineWrapperFunc(log)
)

func init() {
	rpcclient.UseLogger(log)
}
