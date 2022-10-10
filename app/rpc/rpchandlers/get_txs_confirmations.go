package rpchandlers

import (
	"errors"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetTxsConfirmations handles the respectively named RPC command
func HandleGetTxsConfirmations(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	var err error

	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetTxsConfirmationsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getTxsConfirmationsRequest := request.(*appmessage.GetTxsConfirmationsRequestMessage)

	domainTxIDs := make([]*externalapi.DomainTransactionID, len(getTxsConfirmationsRequest.TxIDs))

	for i := range domainTxIDs {
		domainTxIDs[i], err = externalapi.NewDomainTransactionIDFromString(getTxsConfirmationsRequest.TxIDs[i])
		if err != nil {
			errorMessage := &appmessage.GetTxsConfirmationsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("error parsing txID: %s", getTxsConfirmationsRequest.TxIDs[i])
			return errorMessage, nil
		}
	}

	txIDsToConfirmations, _, err := context.TXIndex.GetTXsConfirmations(domainTxIDs)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxsConfirmationsResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	txIDConfirmationPairs := make([]*appmessage.TxIDConfirmationsPair, len(txIDsToConfirmations))

	i := 0
	for txID, Confirmations := range txIDsToConfirmations {
		txIDConfirmationPairs[i] = &appmessage.TxIDConfirmationsPair{
			TxID:          txID.String(),
			Confirmations: Confirmations,
		}
		i++
	}

	response := appmessage.NewGetTxsConfirmationsResponse(txIDConfirmationPairs)

	return response, nil
}
