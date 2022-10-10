package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetTxConfirmations sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetTxConfirmations(txID string) (*appmessage.GetTxConfirmationsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetTxConfirmationsRequest(txID))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetTxConfirmationsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getTxConfirmationsResponse := response.(*appmessage.GetTxConfirmationsResponseMessage)
	if getTxConfirmationsResponse.Error != nil {
		return nil, c.convertRPCError(getTxConfirmationsResponse.Error)
	}
	return getTxConfirmationsResponse, nil
}
