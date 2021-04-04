package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetInfo sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetInfo() (*appmessage.GetInfoResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetInfoRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetInfoResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getInfoResponse := response.(*appmessage.GetInfoResponseMessage)
	if getInfoResponse.Error != nil {
		return nil, c.convertRPCError(getInfoResponse.Error)
	}
	return getInfoResponse, nil
}
