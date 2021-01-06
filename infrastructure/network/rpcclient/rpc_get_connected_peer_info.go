package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetConnectedPeerInfo sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetConnectedPeerInfo() (*appmessage.GetConnectedPeerInfoResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetConnectedPeerInfoRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetConnectedPeerInfoResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getConnectedPeerInfoResponse := response.(*appmessage.GetConnectedPeerInfoResponseMessage)
	if getConnectedPeerInfoResponse.Error != nil {
		return nil, c.convertRPCError(getConnectedPeerInfoResponse.Error)
	}
	return getConnectedPeerInfoResponse, nil
}
