package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetAcceptingBlockHashOfTx sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetAcceptingBlockHashOfTx(txID string) (*appmessage.GetAcceptingBlockHashOfTxResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetAcceptingBlockHashOfTxRequest(txID))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetAcceptingBlockHashOfTxResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getAcceptingBlockHashOfTxResponse := response.(*appmessage.GetAcceptingBlockHashOfTxResponseMessage)
	if getAcceptingBlockHashOfTxResponse.Error != nil {
		return nil, c.convertRPCError(getAcceptingBlockHashOfTxResponse.Error)
	}
	return getAcceptingBlockHashOfTxResponse, nil
}
