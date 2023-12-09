package rpcclient

import "github.com/zoomy-network/zoomyd/app/appmessage"

// GetBlock sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetBlock(hash string, includeTransactions bool) (
	*appmessage.GetBlockResponseMessage, error) {

	err := c.rpcRouter.outgoingRoute().Enqueue(
		appmessage.NewGetBlockRequestMessage(hash, includeTransactions))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlockResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetBlockResponse := response.(*appmessage.GetBlockResponseMessage)
	if GetBlockResponse.Error != nil {
		return nil, c.convertRPCError(GetBlockResponse.Error)
	}
	return GetBlockResponse, nil
}
