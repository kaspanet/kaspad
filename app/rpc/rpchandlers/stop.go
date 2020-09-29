package rpchandlers

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const secondsBeforeStop = 5

// HandleStop handles the respectively named RPC command
func HandleStop(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	log.Warnf("Stop RPC called... Waiting for %d seconds before shutting down Kaspad", secondsBeforeStop)

	spawn("Stop", func() {
		<-time.After(5 * time.Second)
		close(context.StopChan)
	})

	response := appmessage.NewStopResponseMessage()
	return response, nil
}
