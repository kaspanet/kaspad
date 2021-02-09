package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// SubmitBlock sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) SubmitBlock(block *externalapi.DomainBlock) (appmessage.RejectReason, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(
		appmessage.NewSubmitBlockRequestMessage(appmessage.DomainBlockToMsgBlock(block)))
	if err != nil {
		return appmessage.RejectReasonNone, err
	}
	response, err := c.route(appmessage.CmdSubmitBlockResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return appmessage.RejectReasonNone, err
	}
	submitBlockResponse := response.(*appmessage.SubmitBlockResponseMessage)
	if submitBlockResponse.Error != nil {
		return submitBlockResponse.RejectReason, c.convertRPCError(submitBlockResponse.Error)
	}
	return appmessage.RejectReasonNone, nil
}
