package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) GetPruningWindowRoots() (*appmessage.GetPruningWindowRootsResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetPeerAddressesRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetPruningWindowRootsRequestMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	convertedResp := response.(*appmessage.GetPruningWindowRootsResponseMessage)
	if convertedResp.Error != nil {
		return nil, c.convertRPCError(convertedResp.Error)
	}
	return convertedResp, nil
}
