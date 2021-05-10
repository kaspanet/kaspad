package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleEstimateNetworkHashesPerSecond handles the respectively named RPC command
func HandleEstimateNetworkHashesPerSecond(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	estimateNetworkHashesPerSecondRequest := request.(*appmessage.EstimateNetworkHashesPerSecondRequestMessage)

	windowSize := int(estimateNetworkHashesPerSecondRequest.WindowSize)
	networkHashesPerSecond, err := context.Domain.Consensus().EstimateNetworkHashesPerSecond(windowSize)
	if err != nil {
		response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
		response.Error = appmessage.RPCErrorf("could not resolve network hashes per "+
			"second for window size %d: %s", windowSize, err)
		return response, nil
	}

	return appmessage.NewEstimateNetworkHashesPerSecondResponseMessage(networkHashesPerSecond), nil
}
