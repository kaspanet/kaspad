package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleGetBalancesByAddresses handles the respectively named RPC command
func HandleGetBalancesByAddresses(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	if !context.Config.UTXOIndex {
		errorMessage := &appmessage.GetBalancesByAddressesResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Method unavailable when kaspad is run without --utxoindex")
		return errorMessage, nil
	}

	getBalancesByAddressesRequest := request.(*appmessage.GetBalancesByAddressesRequestMessage)

	allEntries := make([]*appmessage.BalancesByAddressesEntry, len(getBalancesByAddressesRequest.Addresses))
	for i, address := range getBalancesByAddressesRequest.Addresses {
		balance, err := getBalanceByAddress(context, address)

		if err != nil {
			rpcError := &appmessage.RPCError{}
			if !errors.As(err, rpcError) {
				return nil, err
			}
			errorMessage := &appmessage.GetUTXOsByAddressesResponseMessage{}
			errorMessage.Error = rpcError
			return errorMessage, nil
		}
		allEntries[i] = &appmessage.BalancesByAddressesEntry{
			Address: address,
			Balance: balance,
		}
	}

	response := appmessage.NewGetBalancesByAddressesResponse(allEntries)
	return response, nil
}
