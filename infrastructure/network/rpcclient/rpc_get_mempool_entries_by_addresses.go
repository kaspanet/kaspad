package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetMempoolEntriesByAddresses sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetMempoolEntriesByAddresses(addresses []string) (*appmessage.GetMempoolEntriesByAddressesResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetMempoolEntriesByAddressesRequestMessage(addresses))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetMempoolEntriesByAddressesResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getMempoolEntriesByAddressesResponse := response.(*appmessage.GetMempoolEntriesByAddressesResponseMessage)
	if getMempoolEntriesByAddressesResponse.Error != nil {
		return nil, c.convertRPCError(getMempoolEntriesByAddressesResponse.Error)
	}
	return getMempoolEntriesByAddressesResponse, nil
}
