package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetMempoolEntries sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetMempoolEntries() (*appmessage.GetMempoolEntriesResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetMempoolEntriesRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetMempoolEntriesResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getMempoolEntriesResponse := response.(*appmessage.GetMempoolEntriesResponseMessage)
	if getMempoolEntriesResponse.Error != nil {
		return nil, c.convertRPCError(getMempoolEntriesResponse.Error)
	}
	return getMempoolEntriesResponse, nil
}
