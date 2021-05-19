package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetBlocks sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetBlocks(lowHash string, includeBlocks bool,
	includeTransactions bool) (*appmessage.GetBlocksResponseMessage, error) {

	err := c.rpcRouter.outgoingRoute().Enqueue(
		appmessage.NewGetBlocksRequestMessage(lowHash, includeBlocks, includeTransactions))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlocksResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetBlocksResponse := response.(*appmessage.GetBlocksResponseMessage)
	if GetBlocksResponse.Error != nil {
		return nil, c.convertRPCError(GetBlocksResponse.Error)
	}
	return GetBlocksResponse, nil
}
