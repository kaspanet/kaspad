package jsonrpc

import (
	"github.com/kaspanet/kaspad/kasparov/logger"
	"github.com/kaspanet/kaspad/rpcclient"
	"github.com/kaspanet/kaspad/util/panics"
)

var (
	log   = logger.BackendLog.Logger("RPCC")
	spawn = panics.GoroutineWrapperFunc(log)
)

func init() {
	rpcclient.UseLogger(log)
}
