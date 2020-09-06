package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

func (c *RPCClient) GetChainFromBlock(startHash string, includeBlockVerboseData bool) (*appmessage.GetChainFromBlockResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetChainFromBlockRequestMessage(startHash, includeBlockVerboseData))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetChainFromBlockResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	GetChainFromBlockResponse := response.(*appmessage.GetChainFromBlockResponseMessage)
	if GetChainFromBlockResponse.Error != nil {
		return nil, c.convertRPCError(GetChainFromBlockResponse.Error)
	}
	return GetChainFromBlockResponse, nil
}
