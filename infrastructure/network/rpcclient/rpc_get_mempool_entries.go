package rpcclient

import "github.com/zoomy-network/zoomyd/app/appmessage"

// GetMempoolEntries sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetMempoolEntries(includeOrphanPool bool, filterTransactionPool bool) (*appmessage.GetMempoolEntriesResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetMempoolEntriesRequestMessage(includeOrphanPool, filterTransactionPool))
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
