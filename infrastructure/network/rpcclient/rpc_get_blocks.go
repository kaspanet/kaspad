package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) GetBlocks(lowHash string, includeBlockHexes bool,
	includeBlockVerboseData bool) (*appmessage.GetBlocksResponseMessage, error) {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetBlocksRequestMessage(lowHash, includeBlockHexes, includeBlockVerboseData))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetBlocksResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetBlocksResponse := response.(*appmessage.GetBlocksResponseMessage)
	if GetBlocksResponse.Error != nil {
		return nil, c.convertRPCError(GetBlocksResponse.Error)
	}
	return GetBlocksResponse, nil
}
