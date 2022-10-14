package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetIncludingBlocksTxs sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetIncludingBlocksTxs(txIDs []string, includeTransactions bool) (*appmessage.GetIncludingBlocksOfTxsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetIncludingBlocksOfTxsRequest(txIDs, includeTransactions))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetIncludingBlocksOfTxsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getIncludingBlocksOfTxsResponse := response.(*appmessage.GetIncludingBlocksOfTxsResponseMessage)
	if getIncludingBlocksOfTxsResponse.Error != nil {
		return nil, c.convertRPCError(getIncludingBlocksOfTxsResponse.Error)
	}
	return getIncludingBlocksOfTxsResponse, nil
}
