package rpcclient

import (
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

// SubmitTransaction sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) SubmitTransaction(transaction *appmessage.RPCTransaction, allowOrphan bool) (*appmessage.SubmitTransactionResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewSubmitTransactionRequestMessage(transaction, allowOrphan))
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdSubmitTransactionResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	submitTransactionResponse := response.(*appmessage.SubmitTransactionResponseMessage)
	if submitTransactionResponse.Error != nil {
		return nil, c.convertRPCError(submitTransactionResponse.Error)
	}

	return submitTransactionResponse, nil
}
