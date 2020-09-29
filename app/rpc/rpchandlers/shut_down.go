package rpchandlers

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const pauseBeforeShutDown = time.Second

// HandleShutDown handles the respectively named RPC command
func HandleShutDown(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	log.Warn("Stop RPC called.")

	// Wait a second before stopping, to allow time to return the response to the caller
	spawn("ShutDown", func() {
		<-time.After(pauseBeforeShutDown)
		close(context.ShutDownChan)
	})

	response := appmessage.NewShutDownResponseMessage()
	return response, nil
}
