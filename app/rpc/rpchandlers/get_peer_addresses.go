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
	addressMessages := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(netAddresses))
	for i, netAddress := range netAddresses {
		addressWithPort := net.JoinHostPort(netAddress.IP.String(), strconv.FormatUint(uint64(netAddress.Port), 10))
		addressMessages[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: addressWithPort}
	}

	bannedAddresses := context.AddressManager.BannedAddresses()
	bannedAddressMessages := make([]*appmessage.GetPeerAddressesKnownAddressMessage, len(bannedAddresses))
	for i, netAddress := range bannedAddresses {
		addressWithPort := net.JoinHostPort(netAddress.IP.String(), strconv.FormatUint(uint64(netAddress.Port), 10))
		bannedAddressMessages[i] = &appmessage.GetPeerAddressesKnownAddressMessage{Addr: addressWithPort}
	}

	response := appmessage.NewGetPeerAddressesResponseMessage(addressMessages, bannedAddressMessages)
	return response, nil
}
