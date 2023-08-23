package rpcclient

import "github.com/c4ei/yunseokyeol/app/appmessage"

// GetMempoolEntry sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetMempoolEntry(txID string, includeOrphanPool bool, filterTransactionPool bool) (*appmessage.GetMempoolEntryResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetMempoolEntryRequestMessage(txID, includeOrphanPool, filterTransactionPool))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetMempoolEntryResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getMempoolEntryResponse := response.(*appmessage.GetMempoolEntryResponseMessage)
	if getMempoolEntryResponse.Error != nil {
		return nil, c.convertRPCError(getMempoolEntryResponse.Error)
	}
	return getMempoolEntryResponse, nil
}
