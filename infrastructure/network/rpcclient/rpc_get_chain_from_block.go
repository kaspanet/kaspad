package rpcclient

import "github.com/c4ei/yunseokyeol/app/appmessage"

// GetVirtualSelectedParentChainFromBlock sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetVirtualSelectedParentChainFromBlock(startHash string, includeAcceptedTransactionIDs bool) (
	*appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(
		appmessage.NewGetVirtualSelectedParentChainFromBlockRequestMessage(startHash, includeAcceptedTransactionIDs))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetVirtualSelectedParentChainFromBlockResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetVirtualSelectedParentChainFromBlockResponse := response.(*appmessage.GetVirtualSelectedParentChainFromBlockResponseMessage)
	if GetVirtualSelectedParentChainFromBlockResponse.Error != nil {
		return nil, c.convertRPCError(GetVirtualSelectedParentChainFromBlockResponse.Error)
	}
	return GetVirtualSelectedParentChainFromBlockResponse, nil
}
