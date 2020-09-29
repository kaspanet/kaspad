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
	log.Warn("Stop RPC called.", secondsBeforeStop)

	// Wait a few seconds before stopping, to allow time to return the response to the caller
	spawn("Stop", func() {
		<-time.After(5 * time.Second)
		close(context.StopChan)
	})

	response := appmessage.NewStopResponseMessage()
	return response, nil
}
