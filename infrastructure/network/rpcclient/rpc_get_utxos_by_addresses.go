package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetUTXOsByAddresses sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetUTXOsByAddresses(addresses []string) (*appmessage.GetUTXOsByAddressesResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetUTXOsByAddressesRequestMessage(addresses))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetUTXOsByAddressesResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getUTXOsByAddressesResponse := response.(*appmessage.GetUTXOsByAddressesResponseMessage)
	if getUTXOsByAddressesResponse.Error != nil {
		return nil, c.convertRPCError(getUTXOsByAddressesResponse.Error)
	}
	return getUTXOsByAddressesResponse, nil
}
