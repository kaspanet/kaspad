package rpcclient

import "github.com/c4ei/YunSeokYeol/app/appmessage"

// Ban sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) Ban(ip string) (*appmessage.BanResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewBanRequestMessage(ip))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdBanRequestMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	banResponse := response.(*appmessage.BanResponseMessage)
	if banResponse.Error != nil {
		return nil, c.convertRPCError(banResponse.Error)
	}
	return banResponse, nil
}
