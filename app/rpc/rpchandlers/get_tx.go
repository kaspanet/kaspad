package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetTx handles the respectively named RPC command
func HandleGetTx(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getTxRequest := request.(*appmessage.GetTxRequestMessage)

	domainTxID, err := externalapi.NewDomainTransactionIDFromString(getTxRequest.TxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	transaction, found, err := context.TXIndex.GetTX(domainTxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}
	if !found {
		errorMessage := &appmessage.GetTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not find tx in the txindex database for txID: %s", domainTxID.String())
		return errorMessage, nil
	}

	blockForVerboseData, found, err := context.TXIndex.TXAcceptingBlock(domainTxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}
	if !found {
		errorMessage := &appmessage.GetTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not find accepting block in the txindex database for txID: %s", domainTxID.String())
		return errorMessage, nil
	}

	rpcTransaction := appmessage.DomainTransactionToRPCTransaction(transaction)
	err = context.PopulateTransactionWithVerboseData(rpcTransaction, blockForVerboseData.Header)
	if err != nil {
		if errors.Is(err, rpccontext.ErrBuildBlockVerboseDataInvalidBlock) {
			errorMessage := &appmessage.GetTxResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Block %s is invalid", consensushashing.BlockHash(blockForVerboseData).String())
			return errorMessage, nil
		}
		return nil, err
	}

	response := appmessage.NewGetTxResponse(rpcTransaction)

	return response, nil
}
