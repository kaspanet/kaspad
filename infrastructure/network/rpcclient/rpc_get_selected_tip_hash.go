package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetSelectedTipHash sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetSelectedTipHash() (*appmessage.GetSelectedTipHashResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetSelectedTipHashRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetSelectedTipHashResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getSelectedTipHashResponse := response.(*appmessage.GetSelectedTipHashResponseMessage)
	if getSelectedTipHashResponse.Error != nil {
		return nil, c.convertRPCError(getSelectedTipHashResponse.Error)
	}
	return getSelectedTipHashResponse, nil
}
