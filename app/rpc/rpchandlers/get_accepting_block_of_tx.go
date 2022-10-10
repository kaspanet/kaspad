package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetAcceptingBlockOfTx handles the respectively named RPC command
func HandleGetAcceptingBlockOfTx(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetAcceptingBlockHashOfTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getAcceptingBlockHashOfTxRequest := request.(*appmessage.GetAcceptingBlockOfTxRequestMessage)

	domainTxID, err := externalapi.NewDomainTransactionIDFromString(getAcceptingBlockHashOfTxRequest.TxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetAcceptingBlockOfTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	acceptingBlock, found, err := context.TXIndex.TXAcceptingBlock(domainTxID)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetAcceptingBlockOfTxResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}
	if !found {
		errorMessage := &appmessage.GetAcceptingBlockOfTxResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not find accepting block in the txindex database for txID: %s", domainTxID.String())
		return errorMessage, nil
	}

	rpcAcceptingBlock := appmessage.DomainBlockToRPCBlock(acceptingBlock)
	err = context.PopulateBlockWithVerboseData(rpcAcceptingBlock, acceptingBlock.Header, acceptingBlock, getAcceptingBlockHashOfTxRequest.IncludeTransactions)
	if err != nil {
		if errors.Is(err, rpccontext.ErrBuildBlockVerboseDataInvalidBlock) {
			errorMessage := &appmessage.GetAcceptingBlockOfTxResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Block %s is invalid", consensushashing.BlockHash(acceptingBlock).String())
			return errorMessage, nil
		}
		return nil, err
	}

	response := appmessage.NewGetAcceptingBlockOfTxResponse(rpcAcceptingBlock)

	return response, nil
}
