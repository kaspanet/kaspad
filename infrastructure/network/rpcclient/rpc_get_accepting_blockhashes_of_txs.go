package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetAcceptingBlockHashesOfTxs sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetAcceptingBlockHashesOfTxs(txIDs []string) (*appmessage.GetAcceptingBlockHashesOfTxsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetAcceptingBlockHashesOfTxsRequest(txIDs))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetAcceptingBlockHashesOfTxsResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getAcceptingBlockHashesOfTxsResponse := response.(*appmessage.GetAcceptingBlockHashesOfTxsResponseMessage)
	if getAcceptingBlockHashesOfTxsResponse.Error != nil {
		return nil, c.convertRPCError(getAcceptingBlockHashesOfTxsResponse.Error)
	}
	return getAcceptingBlockHashesOfTxsResponse, nil
}
