package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetIncludingBlockHashesOfTxs sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetIncludingBlockHashesOfTxs(txIDs []string) (*appmessage.GetIncludingBlockHashesOfTxsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetIncludingBlockHashesOfTxsRequest(txIDs))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetIncludingBlockHashesOfTxsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getIncludingBlockHashesOfTxsResponse := response.(*appmessage.GetIncludingBlockHashesOfTxsResponseMessage)
	if getIncludingBlockHashesOfTxsResponse.Error != nil {
		return nil, c.convertRPCError(getIncludingBlockHashesOfTxsResponse.Error)
	}
	return getIncludingBlockHashesOfTxsResponse, nil
}
