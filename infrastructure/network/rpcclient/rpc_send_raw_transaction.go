package rpcclient

import (
	"strings"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// SubmitTransaction sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) SubmitTransaction(transaction *appmessage.RPCTransaction, transactionID string, allowOrphan bool) (*appmessage.SubmitTransactionResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewSubmitTransactionRequestMessage(transaction, allowOrphan))
	if err != nil {
		return nil, err
	}
	for {
		response, err := c.route(appmessage.CmdSubmitTransactionResponseMessage).DequeueWithTimeout(c.timeout)
		if err != nil {
			return nil, err
		}
		submitTransactionResponse := response.(*appmessage.SubmitTransactionResponseMessage)
		// Match the response to the expected ID. If they are different it means we got an old response which we
		// previously timed-out on, so we log and continue waiting for the correct current response.
		if submitTransactionResponse.TransactionID != transactionID {
			if submitTransactionResponse.Error != nil {
				// A non-updated Kaspad might return an empty ID in the case of error, so in
				// such a case we fallback to checking if the error contains the expected ID
				if submitTransactionResponse.TransactionID != "" || !strings.Contains(submitTransactionResponse.Error.Message, transactionID) {
					log.Warnf("SubmitTransaction: received an error response for previous request: %s", submitTransactionResponse.Error)
					continue
				}

			} else {
				log.Warnf("SubmitTransaction: received a successful response for previous request with ID %s",
					submitTransactionResponse.TransactionID)
				continue
			}
		}
		if submitTransactionResponse.Error != nil {
			return nil, c.convertRPCError(submitTransactionResponse.Error)
		}

		return submitTransactionResponse, nil
	}
}
