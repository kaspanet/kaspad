package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionid"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleGetMempoolEntry handles the respectively named RPC command
func HandleGetMempoolEntry(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {

	transaction := &externalapi.DomainTransaction{}
	var found bool
	var isOrphan bool

	getMempoolEntryRequest := request.(*appmessage.GetMempoolEntryRequestMessage)

	transactionID, err := transactionid.FromString(getMempoolEntryRequest.TxID)
	if err != nil {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transaction ID could not be parsed: %s", err)
		return errorMessage, nil
	}

	transactionPoolTransaction, orphanPoolTransaction, found := context.Domain.MiningManager().GetTransaction(transactionID, !getMempoolEntryRequest.FilterTransactionPool, getMempoolEntryRequest.IncludeOrphanPool)

	if !found {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Transaction %s was not found", transactionID)
		return errorMessage, nil
	}

	if transactionPoolTransaction != nil && orphanPoolTransaction != nil {
		return nil, errors.Errorf("Transaction %s is both an orphan, and not. This should never not happen", transactionID)
	} else if transactionPoolTransaction != nil {
		transaction = transactionPoolTransaction
		isOrphan = false
	} else if orphanPoolTransaction != nil {
		transaction = orphanPoolTransaction
		isOrphan = true
	}

	rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
	err = context.PopulateTransactionWithVerboseData(rpcTransaction, nil)
	if err != nil {
		return nil, err
	}
	return appmessage.NewGetMempoolEntryResponseMessage(transaction.Fee, rpcTransaction, isOrphan), nil
}
