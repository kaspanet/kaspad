package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetUTXOsByAddress handles the respectively named RPC command
func HandleGetUTXOsByAddress(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	// TODO: Rewrite this
/*	getUTXOsRequest := request.(*appmessage.GetUTXOsByAddressRequestMessage)
	outpoints, err := context.DAG.UTXOMap().GetUTXOsByAddress(getUTXOsRequest.Address)
	if err != nil {
		return nil, err
	}

	fullUTXO := context.DAG.UTXOSet()
	utxosVerboseData := make([]*appmessage.UTXOVerboseData, 0, len(outpoints))
	for outpoint, _ := range outpoints {
		utxoEntry, ok := fullUTXO.Get(outpoint)
		if !ok {
			return nil, errors.New("UTXO not found")
		}
		utxoVerboseData := &appmessage.UTXOVerboseData{
			Amount:         utxoEntry.Amount(),
			ScriptPubKey:   utxoEntry.ScriptPubKey(),
			BlockBlueScore: utxoEntry.BlockBlueScore(),
			IsCoinbase:     utxoEntry.IsCoinbase(),
		}
		utxosVerboseData = append(utxosVerboseData, utxoVerboseData)
	}

	response := appmessage.NewGetUTXOsByAddressResponseMessage(getUTXOsRequest.Address, utxosVerboseData)
	return response, nil*/
	return nil, nil
}
