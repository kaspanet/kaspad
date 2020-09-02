package client

import "github.com/kaspanet/kaspad/app/appmessage"

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
