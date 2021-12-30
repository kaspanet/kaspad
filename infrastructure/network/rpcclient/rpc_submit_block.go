package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (c *RPCClient) submitBlock(block *externalapi.DomainBlock, allowNonDAABlocks bool) (appmessage.RejectReason, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(
		appmessage.NewSubmitBlockRequestMessage(appmessage.DomainBlockToRPCBlock(block), allowNonDAABlocks))
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

// SubmitBlock sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) SubmitBlock(block *externalapi.DomainBlock) (appmessage.RejectReason, error) {
	return c.submitBlock(block, false)
}

// SubmitBlockAlsoIfNonDAA operates the same as SubmitBlock with the exception that `allowNonDAABlocks` is set to true
func (c *RPCClient) SubmitBlockAlsoIfNonDAA(block *externalapi.DomainBlock) (appmessage.RejectReason, error) {
	return c.submitBlock(block, true)
}
