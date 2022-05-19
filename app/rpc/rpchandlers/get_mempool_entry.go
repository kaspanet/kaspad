package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetMempoolEntry handles the respectively named RPC command
func HandleGetMempoolEntry(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {

	transaction := &externalapi.DomainTransaction{}
	var ok bool
	var isOrphan bool

	getMempoolEntryRequest := request.(*appmessage.GetMempoolEntryRequestMessage)

	transactionID, err := transactionid.FromString(getMempoolEntryRequest.TxID)
	if err != nil {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transaction ID could not be parsed: %s", err)
		return errorMessage, nil
	}

	if getMempoolEntryRequest.IncludeTransactionPool && getMempoolEntryRequest.IncludeOrphanPool { //both true

		transaction, ok = context.Domain.MiningManager().GetTransaction(transactionID)
		if !ok {
			transaction, ok = context.Domain.MiningManager().GetOrphanTransaction(transactionID)
			if !ok {
				errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
				errorMessage.Error = appmessage.RPCErrorf("Transaction %s was not found", transactionID)
				return errorMessage, nil
			}
			isOrphan = true
		}
		isOrphan = false

	} else if getMempoolEntryRequest.IncludeTransactionPool && !getMempoolEntryRequest.IncludeOrphanPool { //only transactions
		transaction, ok = context.Domain.MiningManager().GetTransaction(transactionID)
		if !ok {
			errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Transaction %s was not found", transactionID)
			return errorMessage, nil
		}
		isOrphan = true

	} else if !getMempoolEntryRequest.IncludeTransactionPool && getMempoolEntryRequest.IncludeOrphanPool { //only orphans
		transaction, ok = context.Domain.MiningManager().GetOrphanTransaction(transactionID)
		if !ok {
			errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Transaction %s was not found", transactionID)
			return errorMessage, nil
		}
		isOrphan = false
	} else if !(getMempoolEntryRequest.IncludeTransactionPool || getMempoolEntryRequest.IncludeOrphanPool) {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Request is not querying any mempool pools")
		return errorMessage, nil

	}
	rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
	err = context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
	if err != nil {
		return nil, err
	}
	return appmessage.NewGetMempoolEntryResponseMessage(transaction.Fee, rpcTransaction, isOrphan), nil
}
