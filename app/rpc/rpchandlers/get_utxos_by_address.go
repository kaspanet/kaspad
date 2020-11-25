package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetUTXOsByAddress handles the respectively named RPC command
func HandleGetUTXOsByAddress(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getUTXOsRequest := request.(*appmessage.GetUTXOsByAddressRequestMessage)
	utxosByAddress, err := context.UTXOAddressIndex.GetUTXOsByAddress(getUTXOsRequest.Address)
	if err != nil {
		return nil, err
	}

	utxosVerboseData := make([]*appmessage.UTXOVerboseData, 0, len(utxosByAddress))
	for outpoint, utxoEntry := range utxosByAddress {
		utxoVerboseData := &appmessage.UTXOVerboseData{
			Amount:         utxoEntry.Amount,
			ScriptPubKey:   utxoEntry.ScriptPublicKey,
			BlockBlueScore: utxoEntry.BlockBlueScore,
			IsCoinbase:     utxoEntry.IsCoinbase,
			TxID:           outpoint.TransactionID.String(),
			Index:          outpoint.Index,
		}
		utxosVerboseData = append(utxosVerboseData, utxoVerboseData)
	}

	response := appmessage.NewGetUTXOsByAddressResponseMessage(getUTXOsRequest.Address, utxosVerboseData)
	return response, nil
}
