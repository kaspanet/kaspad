package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"

	"github.com/pkg/errors"
)

// HandleGetAcceptingBlockHashesOfTxs handles the respectively named RPC command
func HandleGetAcceptingBlockHashesOfTxs(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	var err error

	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetAcceptingBlockHashesOfTxsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getAcceptingBlockHashesOfTxsRequest := request.(*appmessage.GetAcceptingBlockHashesOfTxsRequestMessage)

	domainTxIDs := make([]*externalapi.DomainTransactionID, len(getAcceptingBlockHashesOfTxsRequest.TxIDs))
	for i := range domainTxIDs {
		domainTxIDs[i], err = externalapi.NewDomainTransactionIDFromString(getAcceptingBlockHashesOfTxsRequest.TxIDs[i])
		if err != nil {
			errorMessage := &appmessage.GetAcceptingBlockHashesOfTxsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("error parsing txID: %s", getAcceptingBlockHashesOfTxsRequest.TxIDs[i])
			return errorMessage, nil
		}
	}
	acceptingBlockHashes, _, err := context.TXIndex.TXAcceptingBlockHashes(domainTxIDs)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	txIDBlockHashpairs := make([]*appmessage.TxIDBlockHashPair, len(acceptingBlockHashes))
	i := 0
	for txID, blockHash := range acceptingBlockHashes {
		txIDBlockHashpairs[i] = &appmessage.TxIDBlockHashPair{
			TxID: txID.String(),
			Hash: blockHash.String(),
		}
		i++
	}

	response := appmessage.NewGetAcceptingBlockHashesOfTxsResponse(txIDBlockHashpairs)

	return response, nil
}
