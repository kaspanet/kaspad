package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) GetBlockHex(hash string, subnetworkID string) (*appmessage.GetBlockHexResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBlockHexRequestMessage(hash, subnetworkID))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlockHexResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	getBlockHexResponse := response.(*appmessage.GetBlockHexResponseMessage)
	if getBlockHexResponse.Error != nil {
		return nil, c.convertRPCError(getBlockHexResponse.Error)
	}
	return getBlockHexResponse, nil
}
