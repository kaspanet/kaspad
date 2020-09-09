package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetBlockDAGInfo sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetBlockDAGInfo() (*appmessage.GetBlockDAGInfoResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBlockDAGInfoRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlockDAGInfoResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetBlockDAGInfoResponse := response.(*appmessage.GetBlockDAGInfoResponseMessage)
	if GetBlockDAGInfoResponse.Error != nil {
		return nil, c.convertRPCError(GetBlockDAGInfoResponse.Error)
	}
	return GetBlockDAGInfoResponse, nil
}
