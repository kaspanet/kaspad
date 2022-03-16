package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntry handles the respectively named RPC command
func HandleGetMempoolEntry(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getMempoolEntryRequest := request.(*appmessage.GetMempoolEntryRequestMessage)

	transactionID, err := transactionid.FromString(getMempoolEntryRequest.TxID)
	if err != nil {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transactions ID could not be parsed: %s", err)
		return errorMessage, nil
	}

	transaction, ok := context.Domain.MiningManager().GetTransaction(transactionID)
	if !ok {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transactions %s was not found", transactionID)
		return errorMessage, nil
	}
	rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
	err = context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
	if err != nil {
		return nil, err
	}

	return appmessage.NewGetMempoolEntryResponseMessage(transaction.Fee, rpcTransaction), nil
}
