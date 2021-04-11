package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntries handles the respectively named RPC command
func HandleGetMempoolEntries(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	transactions := context.Domain.MiningManager().AllTransactions()
	entries := make([]*appmessage.MempoolEntry, 0, len(transactions))
	for _, transaction := range transactions {
		rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
		err := context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &appmessage.MempoolEntry{
			Fee:         transaction.Fee,
			Transaction: rpcTransaction,
		})
	}

	return appmessage.NewGetMempoolEntriesResponseMessage(entries), nil
}
