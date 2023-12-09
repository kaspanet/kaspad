package rpcclient

import "github.com/zoomy-network/zoomyd/app/appmessage"

// Unban sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) Unban(ip string) (*appmessage.UnbanResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewUnbanRequestMessage(ip))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdUnbanRequestMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	unbanResponse := response.(*appmessage.UnbanResponseMessage)
	if unbanResponse.Error != nil {
		return nil, c.convertRPCError(unbanResponse.Error)
	}
	return unbanResponse, nil
}
