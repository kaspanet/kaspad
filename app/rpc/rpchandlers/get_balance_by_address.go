package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleGetBalanceByAddress handles the respectively named RPC command
func HandleGetBalanceByAddress(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --utxoindex")
		return errorMessage, nil
	}

	getBalanceByAddressRequest := request.(*appmessage.GetBalanceByAddressRequestMessage)

	var balance uint64 = 0
	addressString := getBalanceByAddressRequest.Address

	address, err := util.DecodeAddress(addressString, context.Config.ActiveNetParams.Prefix)
	if err != nil {
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could decode address '%s': %s", addressString, err)
		return errorMessage, nil
	}

	scriptPublicKey, err := txscript.PayToAddrScript(address)
	if err != nil {
		errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not create a scriptPublicKey for address '%s': %s", addressString, err)
		return errorMessage, nil
	}
	utxoOutpointEntryPairs, err := context.UTXOIndex.UTXOs(scriptPublicKey)
	if err != nil {
		return nil, err
	}
	for _, utxoOutpointEntryPair := range utxoOutpointEntryPairs {
		balance += utxoOutpointEntryPair.Amount()
	}

	response := appmessage.NewGetBalanceByAddressResponse(balance)
	return response, nil
}
