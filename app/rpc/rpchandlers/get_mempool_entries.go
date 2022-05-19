package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntries handles the respectively named RPC command
func HandleGetMempoolEntries(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getMempoolEntriesRequest := request.(*appmessage.GetMempoolEntriesRequestMessage)

	var entries []*appmessage.MempoolEntry
	var err error

	if getMempoolEntriesRequest.IncludeTransactionPool && getMempoolEntriesRequest.IncludeOrphanPool { //both true

		transactionPoolEntries, err := getTransactionPoolMempoolEntries(context)
		if err != nil {
			return nil, err
		}
		orphanPoolEntries, err := getOrphanPoolMempoolEntries(context)
		if err != nil {
			return nil, err
		}
		entries = append(transactionPoolEntries, orphanPoolEntries...)

	} else if getMempoolEntriesRequest.IncludeTransactionPool && !(getMempoolEntriesRequest.IncludeOrphanPool) { //only transactions
		entries, err = getTransactionPoolMempoolEntries(context)
		if err != nil {
			return nil, err
		}

	} else if !(getMempoolEntriesRequest.IncludeTransactionPool) && getMempoolEntriesRequest.IncludeOrphanPool { //only orphans
		entries, err = getOrphanPoolMempoolEntries(context)
		if err != nil {
			return nil, err
		}
	} else if !(getMempoolEntriesRequest.IncludeTransactionPool || getMempoolEntriesRequest.IncludeOrphanPool) {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Request is not querying any mempool pools")
		return errorMessage, nil
	}

	return appmessage.NewGetMempoolEntriesResponseMessage(entries), nil
}

func getTransactionPoolMempoolEntries(context *rpccontext.Context) ([]*appmessage.MempoolEntry, error) {
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
			IsOrphan:    false,
		})
	}
	return entries, nil
}

func getOrphanPoolMempoolEntries(context *rpccontext.Context) ([]*appmessage.MempoolEntry, error) {
	orphanTransactions := context.Domain.MiningManager().AllOrphanTransactions()
	entries := make([]*appmessage.MempoolEntry, 0, len(orphanTransactions))
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
	return entries, nil
}
