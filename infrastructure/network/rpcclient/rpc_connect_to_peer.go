package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// ConnectToPeer sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) ConnectToPeer(address string, isPermanent bool) error {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewConnectToPeerRequestMessage(address, isPermanent))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdConnectToPeerResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	getMempoolEntryResponse := response.(*appmessage.ConnectToPeerResponseMessage)
	if getMempoolEntryResponse.Error != nil {
		return c.convertRPCError(getMempoolEntryResponse.Error)
	}
	return nil
}
