package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetAcceptingBlocksOfTx handles the respectively named RPC command
func HandleGetAcceptingBlocksOfTx(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	var err error

	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetAcceptingBlocksOfTxsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getAcceptingBlocksOfTxsRequest := request.(*appmessage.GetAcceptingBlocksOfTxsRequestMessage)

	domainTxIDs := make([]*externalapi.DomainTransactionID, len(getAcceptingBlocksOfTxsRequest.TxIDs))
	for i := range domainTxIDs {
		domainTxIDs[i], err = externalapi.NewDomainTransactionIDFromString(getAcceptingBlocksOfTxsRequest.TxIDs[i])
		if err != nil {
			errorMessage := &appmessage.GetAcceptingBlocksOfTxsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("error parsing txID: %s, %s", getAcceptingBlocksOfTxsRequest.TxIDs[i], err.Error())
			return errorMessage, nil
		}
	}
	acceptingBlockHashes, _, err := context.TXIndex.TXAcceptingBlocks(domainTxIDs)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	txIDBlockPairs := make([]*appmessage.TxIDBlockPair, len(acceptingBlockHashes))
	i := 0
	for txID, acceptingBlock := range acceptingBlockHashes {
		rpcAcceptingBlock := appmessage.DomainBlockToRPCBlock(acceptingBlock)
		err = context.PopulateBlockWithVerboseData(rpcAcceptingBlock, acceptingBlock.Header, acceptingBlock, getAcceptingBlocksOfTxsRequest.IncludeTransactions)
		if err != nil {
			if errors.Is(err, rpccontext.ErrBuildBlockVerboseDataInvalidBlock) {
				errorMessage := &appmessage.GetAcceptingBlockOfTxResponseMessage{}
				errorMessage.Error = appmessage.RPCErrorf("Block %s is invalid", consensushashing.BlockHash(acceptingBlock).String())
				return errorMessage, nil
			}
			return nil, err
		}
		txIDBlockPairs[i] = &appmessage.TxIDBlockPair{
			TxID:  txID.String(),
			Block: rpcAcceptingBlock,
		}
		i++
	}

	response := appmessage.NewGetAcceptingBlocksOfTxsResponse(txIDBlockPairs)

	return response, nil
}
