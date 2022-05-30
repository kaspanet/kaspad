package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetInfo sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetCoinSupply() (*appmessage.GetCoinSupplyResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetCoinSupplyRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetCoinSupplyResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	geCoinSupplyResponse := response.(*appmessage.GetCoinSupplyResponseMessage)
	if geCoinSupplyResponse.Error != nil {
		return nil, c.convertRPCError(geCoinSupplyResponse.Error)
	}
	return geCoinSupplyResponse, nil
}
