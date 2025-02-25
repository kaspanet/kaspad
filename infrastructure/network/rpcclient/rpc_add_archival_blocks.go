package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) AddArchivalBlocks(blocks []*appmessage.ArchivalBlock) (*appmessage.AddArchivalBlocksResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewAddArchivalBlocksRequestMessage(blocks))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdAddArchivalBlocksRequestMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	convertedResp := response.(*appmessage.AddArchivalBlocksResponseMessage)
	if convertedResp.Error != nil {
		return nil, c.convertRPCError(convertedResp.Error)
	}
	return convertedResp, nil
}
