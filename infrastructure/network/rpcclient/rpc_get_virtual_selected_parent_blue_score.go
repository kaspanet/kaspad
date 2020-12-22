package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetVirtualSelectedParentBlueScore sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetVirtualSelectedParentBlueScore() (*appmessage.GetVirtualSelectedParentBlueScoreResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetVirtualSelectedParentBlueScoreRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetVirtualSelectedParentBlueScoreResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getVirtualSelectedParentBlueScoreResponse := response.(*appmessage.GetVirtualSelectedParentBlueScoreResponseMessage)
	if getVirtualSelectedParentBlueScoreResponse.Error != nil {
		return nil, c.convertRPCError(getVirtualSelectedParentBlueScoreResponse.Error)
	}
	return getVirtualSelectedParentBlueScoreResponse, nil
}
