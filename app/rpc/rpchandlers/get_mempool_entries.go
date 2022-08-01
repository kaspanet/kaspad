package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntries handles the respectively named RPC command
func HandleGetMempoolEntries(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getMempoolEntriesRequest := request.(*appmessage.GetMempoolEntriesRequestMessage)

	entries := make([]*appmessage.MempoolEntry, 0)

	transactionPoolTransactions, orphanPoolTransactions := context.Domain.MiningManager().AllTransactions(!getMempoolEntriesRequest.FilterTransactionPool, getMempoolEntriesRequest.IncludeOrphanPool)

	if !getMempoolEntriesRequest.FilterTransactionPool {
		for _, transaction := range transactionPoolTransactions {
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
	}
	if getMempoolEntriesRequest.IncludeOrphanPool {
		for _, transaction := range orphanPoolTransactions {
			rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
			err := context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
			if err != nil {
				return nil, err
			}
			entries = append(entries, &appmessage.MempoolEntry{
				Fee:         transaction.Fee,
				Transaction: rpcTransaction,
				IsOrphan:    true,
			})
		}
	}

	return appmessage.NewGetMempoolEntriesResponseMessage(entries), nil
}
