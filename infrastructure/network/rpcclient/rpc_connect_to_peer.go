package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// AddPeer sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) AddPeer(address string, isPermanent bool) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewAddPeerRequestMessage(address, isPermanent))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdAddPeerResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	getMempoolEntryResponse := response.(*appmessage.AddPeerResponseMessage)
	if getMempoolEntryResponse.Error != nil {
		return c.convertRPCError(getMempoolEntryResponse.Error)
	}
	return nil
}
