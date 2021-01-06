package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntries handles the respectively named RPC command
func HandleGetMempoolEntries(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {

	transactions := context.Domain.MiningManager().AllTransactions()
	entries := make([]*appmessage.MempoolEntry, 0, len(transactions))
	for _, tx := range transactions {
		transactionVerboseData, err := context.BuildTransactionVerboseData(
			tx, consensushashing.TransactionID(tx).String(), nil, "")
		if err != nil {
			return nil, err
		}

		entries = append(entries, &appmessage.MempoolEntry{
			Fee:                    tx.Fee,
			TransactionVerboseData: transactionVerboseData,
		})
	}

	return appmessage.NewGetMempoolEntriesResponseMessage(entries), nil
}
