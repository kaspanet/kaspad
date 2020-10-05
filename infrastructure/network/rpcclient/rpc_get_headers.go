package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetHeaders sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetHeaders(startHash string, limit uint64, isAscending bool) (*appmessage.GetHeadersResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetHeadersRequestMessage(startHash, limit, isAscending))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetHeadersResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getHeadersResponse := response.(*appmessage.GetHeadersResponseMessage)
	if getHeadersResponse.Error != nil {
		return nil, c.convertRPCError(getHeadersResponse.Error)
	}
	return getHeadersResponse, nil
}
