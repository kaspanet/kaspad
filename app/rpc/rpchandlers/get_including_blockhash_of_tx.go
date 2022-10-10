package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetIncludingBlockHashOfTx handles the respectively named RPC command
func HandleGetIncludingBlockHashOfTx(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetIncludingBlockHashOfTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getIncludingBlockHashOfTxRequest := request.(*appmessage.GetIncludingBlockHashOfTxRequestMessage)

	domainTxID, err := externalapi.NewDomainTransactionIDFromString(getIncludingBlockHashOfTxRequest.TxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetIncludingBlockHashOfTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	includingBlockHash, found, err := context.TXIndex.TXIncludingBlockHash(domainTxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetIncludingBlockHashOfTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}
	if !found {
		errorMessage := &appmessage.GetAcceptingBlockHashOfTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not find including block hash in the txindex database for txID: %s", domainTxID.String())
		return errorMessage, nil
	}

	response := appmessage.NewGetIncludingBlockHashOfTxResponse(includingBlockHash.String())

	return response, nil
}
