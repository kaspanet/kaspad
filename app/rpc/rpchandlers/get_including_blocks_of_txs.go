package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetIncludingBlocksOfTxs handles the respectively named RPC command
func HandleGetIncludingBlocksOfTxs(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	var err error

	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetIncludingBlocksOfTxsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getIncludingBlocksOfTxsRequest := request.(*appmessage.GetIncludingBlocksOfTxsRequestMessage)

	domainTxIDs := make([]*externalapi.DomainTransactionID, len(getIncludingBlocksOfTxsRequest.TxIDs))
	for i := range domainTxIDs {
		domainTxIDs[i], err = externalapi.NewDomainTransactionIDFromString(getIncludingBlocksOfTxsRequest.TxIDs[i])
		if err != nil {
			errorMessage := &appmessage.GetIncludingBlocksOfTxsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("error parsing txID: %s, %s", getIncludingBlocksOfTxsRequest.TxIDs[i], err.Error())
			return errorMessage, nil
		}
	}

	includingBlockHashes, _, err := context.TXIndex.TXIncludingBlocks(domainTxIDs)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetIncludingBlocksOfTxsResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	txIDBlockPairs := make([]*appmessage.TxIDBlockPair, len(includingBlockHashes))
	i := 0
	for txID, includingBlock := range includingBlockHashes {
		rpcIncludingBlock := appmessage.DomainBlockToRPCBlock(includingBlock)
		err = context.PopulateBlockWithVerboseData(rpcIncludingBlock, includingBlock.Header, includingBlock, getIncludingBlocksOfTxsRequest.IncludeTransactions)
		if err != nil {
			if errors.Is(err, rpccontext.ErrBuildBlockVerboseDataInvalidBlock) {
				errorMessage := &appmessage.GetIncludingBlockOfTxResponseMessage{}
				errorMessage.Error = appmessage.RPCErrorf("Block %s is invalid", consensushashing.BlockHash(includingBlock).String())
				return errorMessage, nil
			}
			return nil, err
		}
		txIDBlockPairs[i] = &appmessage.TxIDBlockPair{
			TxID:  txID.String(),
			Block: rpcIncludingBlock,
		}
		i++
	}

	response := appmessage.NewGetIncludingBlocksOfTxsResponse(txIDBlockPairs)

	return response, nil
}
