package rpcclient

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util"
)

func (c *RPCClient) SubmitBlock(block *util.Block) error {
	blockHex := ""
	if block != nil {
		blockBytes, err := block.Bytes()
		if err != nil {
			return err
		}
		blockHex = hex.EncodeToString(blockBytes)
	}
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewSubmitBlockRequestMessage(blockHex))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdSubmitBlockResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	submitBlockResponse := response.(*appmessage.SubmitBlockResponseMessage)
	if submitBlockResponse.Error != nil {
		return c.convertRPCError(submitBlockResponse.Error)
	}
	return nil
}
