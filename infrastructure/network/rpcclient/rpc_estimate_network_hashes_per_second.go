package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// EstimateNetworkHashesPerSecond sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) EstimateNetworkHashesPerSecond(startHash string, windowSize uint32) (*appmessage.EstimateNetworkHashesPerSecondResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewEstimateNetworkHashesPerSecondRequestMessage(startHash, windowSize))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdEstimateNetworkHashesPerSecondResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	estimateNetworkHashesPerSecondResponse := response.(*appmessage.EstimateNetworkHashesPerSecondResponseMessage)
	if estimateNetworkHashesPerSecondResponse.Error != nil {
		return nil, c.convertRPCError(estimateNetworkHashesPerSecondResponse.Error)
	}
	return estimateNetworkHashesPerSecondResponse, nil
}
