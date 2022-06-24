package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetAcceptingBlockHashOfTx handles the respectively named RPC command
func HandleGetAcceptingBlockHashOfTx(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetAcceptingBlockHashOfTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getAcceptingBlockHashOfTxRequest := request.(*appmessage.GetAcceptingBlockHashOfTxRequestMessage)

	domainTxID, err := externalapi.NewDomainTransactionIDFromString(getAcceptingBlockHashOfTxRequest.TxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	acceptingBlockHash, found, err := context.TXIndex.TXAcceptingBlockHash(domainTxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}
	if !found {
		errorMessage := &appmessage.GetAcceptingBlockHashOfTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not find accepting block hash in the txindex database for txID: %s", domainTxID.String())
		return errorMessage, nil
	}

	response := appmessage.NewGetAcceptingBlockHashOfTxResponse(acceptingBlockHash.String())

	return response, nil
}
