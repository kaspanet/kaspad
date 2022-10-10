package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetTx sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetTx(txID string) (*appmessage.GetTxResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetTxRequest(txID))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetTxResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getTxResponse := response.(*appmessage.GetTxResponseMessage)
	if getTxResponse.Error != nil {
		return nil, c.convertRPCError(getTxResponse.Error)
	}
	return getTxResponse, nil
}
