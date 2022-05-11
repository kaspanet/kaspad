package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntries handles the respectively named RPC command
func HandleGetMempoolEntries(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	transactions := context.Domain.MiningManager().AllTransactions()
	orphanTransactions := context.Domain.MiningManager().AllOrphanTransactions()
	entries := make([]*appmessage.MempoolEntry, 0, len(transactions)+len(orphanTransactions))
	for _, transaction := range transactions {
		rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
		err := context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &appmessage.MempoolEntry{
			Fee:         transaction.Fee,
			Transaction: rpcTransaction,
			IsOrphan:    false,
		})
	}

	orphanTransactions = context.Domain.MiningManager().AllOrphanTransactions()

	for _, orphanTransaction := range orphanTransactions {
		rpcTransaction := appmessage.DomainTransactionToRPCTransaction(orphanTransaction)
		err := context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &appmessage.MempoolEntry{
			Fee:         orphanTransaction.Fee,
			Transaction: rpcTransaction,
			IsOrphan:    true,
		})
	}

	return appmessage.NewGetMempoolEntriesResponseMessage(entries), nil
}
