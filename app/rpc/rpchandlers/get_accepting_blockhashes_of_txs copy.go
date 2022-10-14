package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetIncludingBlockHashesOfTxs handles the respectively named RPC command
func HandleGetIncludingBlockHashesOfTxs(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	var err error

	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetIncludingBlockHashesOfTxsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getIncludingBlockHashesOfTxsRequest := request.(*appmessage.GetIncludingBlockHashesOfTxsRequestMessage)

	domainTxIDs := make([]*externalapi.DomainTransactionID, len(getIncludingBlockHashesOfTxsRequest.TxIDs))
	for i := range domainTxIDs {
		domainTxIDs[i], err = externalapi.NewDomainTransactionIDFromString(getIncludingBlockHashesOfTxsRequest.TxIDs[i])
		if err != nil {
			errorMessage := &appmessage.GetIncludingBlockHashesOfTxsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("error parsing txID: %s", getIncludingBlockHashesOfTxsRequest.TxIDs[i])
			return errorMessage, nil
		}
	}

	includingBlockHashes, _, err := context.TXIndex.TXIncludingBlockHashes(domainTxIDs)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetIncludingBlockHashesOfTxsResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	txIDBlockHashpairs := make([]*appmessage.TxIDBlockHashPair, len(includingBlockHashes))
	i := 0
	for txID, blockHash := range includingBlockHashes {
		txIDBlockHashpairs[i] = &appmessage.TxIDBlockHashPair{
			TxID: txID.String(),
			Hash: blockHash.String(),
		}
		i++
	}

	response := appmessage.NewGetIncludingBlockHashesOfTxsResponse(txIDBlockHashpairs)

	return response, nil
}
