package rpchandlers

import (
	"errors"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetTxs handles the respectively named RPC command
func HandleGetTxs(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	var err error

	if !context.Config.TXIndex {
		errorMessage := &appmessage.GetTxsResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --txindex")
		return errorMessage, nil
	}

	getTxsRequest := request.(*appmessage.GetTxsRequestMessage)

	domainTxIDs := make([]*externalapi.DomainTransactionID, len(getTxsRequest.TxIDs))

	for i := range domainTxIDs {
		domainTxIDs[i], err = externalapi.NewDomainTransactionIDFromString(getTxsRequest.TxIDs[i])
		if err != nil {
			errorMessage := &appmessage.GetTxsConfirmationsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("error parsing txID: %s", getTxsRequest.TxIDs[i])
			return errorMessage, nil
		}
	}

	transactions, _, err := context.TXIndex.GetTXs(domainTxIDs)
	if err != nil {
		rpcError := &appmessage.RPCError{}
		if !errors.As(err, &rpcError) {
			return nil, err
		}
		errorMessage := &appmessage.GetTxsResponseMessage{}
		errorMessage.Error = rpcError
		return errorMessage, nil
	}

	rpcTransactions := make([]*appmessage.RPCTransaction, len(transactions))

	for i := range transactions {
		rpcTransactions[i] = appmessage.DomainTransactionToRPCTransaction(transactions[i])
		blockForVerboseData, found, err := context.TXIndex.TXAcceptingBlock(consensushashing.TransactionID(transactions[i]))
		if err != nil {
			rpcError := &appmessage.RPCError{}
			if !errors.As(err, &rpcError) {
				return nil, err
			}
			errorMessage := &appmessage.GetTxsResponseMessage{}
			errorMessage.Error = rpcError
			return errorMessage, nil
		}
		if !found {
			errorMessage := &appmessage.GetTxsResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Could not find accepting block in the txindex database for txID: %s", consensushashing.TransactionID(transactions[i]).String())
			return errorMessage, nil
		}

		err = context.PopulateTransactionWithVerboseData(rpcTransactions[i], blockForVerboseData.Header)
		if err != nil {
			if errors.Is(err, rpccontext.ErrBuildBlockVerboseDataInvalidBlock) {
				errorMessage := &appmessage.GetTxsResponseMessage{}
				errorMessage.Error = appmessage.RPCErrorf("Block %s is invalid", consensushashing.BlockHash(blockForVerboseData).String())
				return errorMessage, nil
			}
			return nil, err
		}
	}

	response := appmessage.NewGetTxsResponse(rpcTransactions)

	return response, nil
}
