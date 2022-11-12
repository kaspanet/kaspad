package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetTxs sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetTxs(txIDs []string) (*appmessage.GetTxsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetTxsRequest(txIDs))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetTxsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getTxsResponse := response.(*appmessage.GetTxsResponseMessage)
	if getTxsResponse.Error != nil {
		return nil, c.convertRPCError(getTxsResponse.Error)
	}
	return getTxsResponse, nil
}
