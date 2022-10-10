package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetAcceptingBlockOfTx sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetAcceptingBlockOfTx(txID string, includeTransactions bool) (*appmessage.GetAcceptingBlockOfTxResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetAcceptingBlockOfTxRequest(txID, includeTransactions))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetAcceptingBlockOfTxResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getAcceptingBlockOfTxResponse := response.(*appmessage.GetAcceptingBlockOfTxResponseMessage)
	if getAcceptingBlockOfTxResponse.Error != nil {
		return nil, c.convertRPCError(getAcceptingBlockOfTxResponse.Error)
	}
	return getAcceptingBlockOfTxResponse, nil
}
