package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetAcceptingBlocksTxs sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetAcceptingBlocksTxs(txIDs []string) (*appmessage.GetAcceptingBlocksOfTxsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetAcceptingBlocksOfTxsRequest(txIDs))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetAcceptingBlocksOfTxsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getAcceptingBlocksOfTxsResponse := response.(*appmessage.GetAcceptingBlocksOfTxsResponseMessage)
	if getAcceptingBlocksOfTxsResponse.Error != nil {
		return nil, c.convertRPCError(getAcceptingBlocksOfTxsResponse.Error)
	}
	return getAcceptingBlocksOfTxsResponse, nil
}
