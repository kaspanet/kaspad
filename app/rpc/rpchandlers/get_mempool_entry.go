package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
)

// HandleGetMempoolEntry handles the respectively named RPC command
func HandleGetMempoolEntry(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getMempoolEntryRequest := request.(*appmessage.GetMempoolEntryRequestMessage)
	txID, err := daghash.NewTxIDFromStr(getMempoolEntryRequest.TxID)
	if err != nil {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse txId: %s", err)
		return errorMessage, nil
	}

	txDesc, ok := context.Mempool.FetchTxDesc(txID)
	if !ok {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("transaction is not in the pool")
		return errorMessage, nil
	}

	transactionVerboseData, err := context.BuildTransactionVerboseData(txDesc.Tx.MsgTx(), txID.String(),
		nil, "", nil, true)
	if err != nil {
		return nil, err
	}

	response := appmessage.NewGetMempoolEntryResponseMessage(txDesc.Fee, transactionVerboseData)
	return response, nil
}
