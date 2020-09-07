package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetPeerAddresses handles the respectively named RPC command
func HandleGetPeerAddresses(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	//TODO: this functionality of AddressManager was removed after refactor. Perhaps, we need to get rid of this code
	//peersState, err := context.AddressManager.PeersStateForSerialization()
	//if err != nil {
	//	return nil, err
	//}
	//addresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(peersState.Addresses))
	//for i, address := range peersState.Addresses {
	//	addresses[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: string(address.Address)}
	//}
	//response := appmessage.NewGetPeerAddressesResponseMessage(addresses)
	//return response, nil
	return nil, nil
}
