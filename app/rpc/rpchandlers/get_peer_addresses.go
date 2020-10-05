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
	netAaddresses := context.AddressManager.Addresses()
	addresses := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(netAaddresses))
	for i, netAddress := range netAaddresses {
		port := strconv.FormatUint(uint64(netAddress.Port), 10)
		addressWithPort := net.JoinHostPort(netAddress.IP.String(), port)
		addresses[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: addressWithPort}
	}
	response := appmessage.NewGetPeerAddressesResponseMessage(addresses)
	return response, nil
}
