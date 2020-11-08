package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetBlockTemplate sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetBlockTemplate(miningAddress string) (*appmessage.GetBlockTemplateResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBlockTemplateRequestMessage(miningAddress))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlockTemplateResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getBlockTemplateResponse := response.(*appmessage.GetBlockTemplateResponseMessage)
	if getBlockTemplateResponse.Error != nil {
		return nil, c.convertRPCError(getBlockTemplateResponse.Error)
	}
	return getBlockTemplateResponse, nil
}
