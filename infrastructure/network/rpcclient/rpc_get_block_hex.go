package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) GetBlock(hash string, subnetworkID string, includeBlockHex bool, includeBlockVerboseData bool) (*appmessage.GetBlockResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBlockRequestMessage(hash, subnetworkID, includeBlockHex, includeBlockVerboseData))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlockResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetBlockResponse := response.(*appmessage.GetBlockResponseMessage)
	if GetBlockResponse.Error != nil {
		return nil, c.convertRPCError(GetBlockResponse.Error)
	}
	return GetBlockResponse, nil
}
