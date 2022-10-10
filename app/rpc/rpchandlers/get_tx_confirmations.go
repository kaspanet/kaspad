package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetTxConfirmations handles the respectively named RPC command
func HandleGetTxConfirmations(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetTxConfirmationsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getTxConfirmationsRequest := request.(*appmessage.GetTxConfirmationsRequestMessage)

	domainTxID, err := externalapi.NewDomainTransactionIDFromString(getTxConfirmationsRequest.TxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxConfirmationsResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	confirmations, _, err := context.TXIndex.GetTXConfirmations(domainTxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxConfirmationsResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	response := appmessage.NewGetTxConfirmationsResponse(confirmations)

	return response, nil
}
