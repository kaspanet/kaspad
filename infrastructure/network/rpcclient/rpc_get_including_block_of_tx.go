package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetIncludingBlockOfTx sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetIncludingBlockOfTx(txID string, includeTransactions bool) (*appmessage.GetIncludingBlockOfTxResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetIncludingBlockOfTxRequest(txID, includeTransactions))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetIncludingBlockOfTxResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getIncludingBlockOfTxResponse := response.(*appmessage.GetIncludingBlockOfTxResponseMessage)
	if getIncludingBlockOfTxResponse.Error != nil {
		return nil, c.convertRPCError(getIncludingBlockOfTxResponse.Error)
	}
	return getIncludingBlockOfTxResponse, nil
}
