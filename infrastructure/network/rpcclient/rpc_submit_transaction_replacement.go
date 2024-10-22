package rpcclient

import (
	"strings"

	"github.com/kaspanet/kaspad/app/appmessage"
)

// SubmitTransactionReplacement sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) SubmitTransactionReplacement(transaction *appmessage.RPCTransaction, transactionID string) (*appmessage.SubmitTransactionReplacementResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewSubmitTransactionReplacementRequestMessage(transaction))
	if err != nil {
		return nil, err
	}
	for {
		response, err := c.route(appmessage.CmdSubmitTransactionReplacementResponseMessage).DequeueWithTimeout(c.timeout)
		if err != nil {
			return nil, err
		}
		SubmitTransactionReplacementResponse := response.(*appmessage.SubmitTransactionReplacementResponseMessage)
		// Match the response to the expected ID. If they are different it means we got an old response which we
		// previously timed-out on, so we log and continue waiting for the correct current response.
		if SubmitTransactionReplacementResponse.TransactionID != transactionID {
			if SubmitTransactionReplacementResponse.Error != nil {
				// A non-updated Kaspad might return an empty ID in the case of error, so in
				// such a case we fallback to checking if the error contains the expected ID
				if SubmitTransactionReplacementResponse.TransactionID != "" || !strings.Contains(SubmitTransactionReplacementResponse.Error.Message, transactionID) {
					log.Warnf("SubmitTransactionReplacement: received an error response for previous request: %s", SubmitTransactionReplacementResponse.Error)
					continue
				}

			} else {
				log.Warnf("SubmitTransactionReplacement: received a successful response for previous request with ID %s",
					SubmitTransactionReplacementResponse.TransactionID)
				continue
			}
		}
		if SubmitTransactionReplacementResponse.Error != nil {
			return nil, c.convertRPCError(SubmitTransactionReplacementResponse.Error)
		}

		return SubmitTransactionReplacementResponse, nil
	}
}
