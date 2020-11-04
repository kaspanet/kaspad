package rpcclient

import (
	"encoding/hex"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// SubmitBlock sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) SubmitBlock(block *externalapi.DomainBlock) error {
	blockBytes, err := block.Bytes()
	if err != nil {
		return err
	}
	blockHex := hex.EncodeToString(blockBytes)
	err = c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewSubmitBlockRequestMessage(blockHex))
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
