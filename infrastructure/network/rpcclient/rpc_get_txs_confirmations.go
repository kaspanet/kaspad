package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetTXsConfirmations sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetTxsConfirmations(txIDs []string) (*appmessage.GetTxsConfirmationsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetTxsConfirmationsRequest(txIDs))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetTxsConfirmationsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getTxsConfirmationsResponse := response.(*appmessage.GetTxsConfirmationsResponseMessage)
	if getTxsConfirmationsResponse.Error != nil {
		return nil, c.convertRPCError(getTxsConfirmationsResponse.Error)
	}
	return getTxsConfirmationsResponse, nil
}

