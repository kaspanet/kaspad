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

	_, ok := context.Mempool.FetchTxDesc(txID)
	if !ok {
		errorMessage := &appmessage.GetMempoolEntryResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("transaction is not in the pool")
		return errorMessage, nil
	}

	response := appmessage.NewGetMempoolEntryResponseMessage()
	return response, nil
}
