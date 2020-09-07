package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/subnetworkid"
)

// HandleGetSubnetwork handles the respectively named RPC command
func HandleGetSubnetwork(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getSubnetworkRequest := request.(*appmessage.GetSubnetworkRequestMessage)

	subnetworkID, err := subnetworkid.NewFromStr(getSubnetworkRequest.SubnetworkID)
	if err != nil {
		errorMessage := &appmessage.GetSubnetworkResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not parse subnetworkID: %s", err)
		return errorMessage, nil
	}

	var gasLimit uint64
	if !subnetworkID.IsBuiltInOrNative() {
		limit, err := context.DAG.GasLimit(subnetworkID)
		if err != nil {
			errorMessage := &appmessage.GetSubnetworkResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Subnetwork %s not found.", subnetworkID)
			return errorMessage, nil
		}
		gasLimit = limit
	}

	response := appmessage.NewGetSubnetworkResponseMessage(gasLimit)
	return response, nil
}
