package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// ResolveFinalityConflict sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) ResolveFinalityConflict(finalityBlockHash string) (*appmessage.ResolveFinalityConflictResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewResolveFinalityConflictRequestMessage(finalityBlockHash))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdResolveFinalityConflictResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	resolveFinalityConflictResponse := response.(*appmessage.ResolveFinalityConflictResponseMessage)
	if resolveFinalityConflictResponse.Error != nil {
		return nil, c.convertRPCError(resolveFinalityConflictResponse.Error)
	}
	return resolveFinalityConflictResponse, nil
}
