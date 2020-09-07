package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetPeerAddresses sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetPeerAddresses() (*appmessage.GetPeerAddressesResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetPeerAddressesRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetPeerAddressesResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getPeerAddressesResponse := response.(*appmessage.GetPeerAddressesResponseMessage)
	if getPeerAddressesResponse.Error != nil {
		return nil, c.convertRPCError(getPeerAddressesResponse.Error)
	}
	return getPeerAddressesResponse, nil
}
