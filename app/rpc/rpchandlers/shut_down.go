package rpchandlers

import (
	"time"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const pauseBeforeShutDown = time.Second

// HandleShutDown handles the respectively named RPC command
func HandleShutDown(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	log.Warn("ShutDown RPC called.")

	// Wait a second before shutting down, to allow time to return the response to the caller
	spawn("HandleShutDown-pauseAndShutDown", func() {
		<-time.After(pauseBeforeShutDown)
		close(context.ShutDownChan)
	})

	response := appmessage.NewShutDownResponseMessage()
	return response, nil
}
