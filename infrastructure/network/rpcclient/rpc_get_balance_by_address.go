package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetBalanceByAddress sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetBalanceByAddress(address string) (*appmessage.GetBalanceByAddressResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBalanceByAddressRequestMessage(address))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBalanceByAddressResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getBalanceByAddressResponse := response.(*appmessage.GetBalanceByAddressResponseMessage)
	if getBalanceByAddressResponse.Error != nil {
		return nil, c.convertRPCError(getBalanceByAddressResponse.Error)
	}
	return getBalanceByAddressResponse, nil
}
