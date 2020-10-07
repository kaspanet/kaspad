package rpchandlers

import (
	"net"
	"strconv"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleGetPeerAddresses handles the respectively named RPC command
func HandleGetPeerAddresses(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	netAddresses := context.AddressManager.Addresses()
	netAddresses = append(netAddresses, context.AddressManager.BannedAddresses()...)

	addresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(netAddresses))
	for i, netAddress := range netAddresses {
		port := strconv.FormatUint(uint64(netAddress.Port), 10)
		addressWithPort := net.JoinHostPort(netAddress.IP.String(), port)
		addresses[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: addressWithPort}
	}
	response := appmessage.NewGetPeerAddressesResponseMessage(addresses)
	return response, nil
}
