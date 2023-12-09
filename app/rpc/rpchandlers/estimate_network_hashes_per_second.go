package rpchandlers

import (
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/zoomy-network/zoomyd/app/rpc/rpccontext"
	"github.com/zoomy-network/zoomyd/domain/consensus/model"
	"github.com/zoomy-network/zoomyd/domain/consensus/model/externalapi"
	"github.com/zoomy-network/zoomyd/infrastructure/network/netadapter/router"
)

// HandleEstimateNetworkHashesPerSecond handles the respectively named RPC command
func HandleEstimateNetworkHashesPerSecond(
	context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {

	estimateNetworkHashesPerSecondRequest := request.(*appmessage.EstimateNetworkHashesPerSecondRequestMessage)

	windowSize := int(estimateNetworkHashesPerSecondRequest.WindowSize)
	startHash := model.VirtualBlockHash
	if estimateNetworkHashesPerSecondRequest.StartHash != "" {
		var err error
		startHash, err = externalapi.NewDomainHashFromString(estimateNetworkHashesPerSecondRequest.StartHash)
		if err != nil {
			response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
			response.Error = appmessage.RPCErrorf("StartHash '%s' is not a valid block hash",
				estimateNetworkHashesPerSecondRequest.StartHash)
			return response, nil
		}
	}

	if context.Config.SafeRPC {
		const windowSizeLimit = 10000
		if windowSize > windowSizeLimit {
			response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
			response.Error =
				appmessage.RPCErrorf(
					"Requested window size %d is larger than max allowed in RPC safe mode (%d)",
					windowSize, windowSizeLimit)
			return response, nil
		}
	}

	if uint64(windowSize) > context.Config.ActiveNetParams.PruningDepth() {
		response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
		response.Error =
			appmessage.RPCErrorf(
				"Requested window size %d is larger than pruning point depth %d",
				windowSize, context.Config.ActiveNetParams.PruningDepth())
		return response, nil
	}

	networkHashesPerSecond, err := context.Domain.Consensus().EstimateNetworkHashesPerSecond(startHash, windowSize)
	if err != nil {
		response := &appmessage.EstimateNetworkHashesPerSecondResponseMessage{}
		response.Error = appmessage.RPCErrorf("could not resolve network hashes per "+
			"second for startHash %s and window size %d: %s", startHash, windowSize, err)
		return response, nil
	}

	return appmessage.NewEstimateNetworkHashesPerSecondResponseMessage(networkHashesPerSecond), nil
}
