package addressexchange

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/network/addressmanager"
	"github.com/kaspanet/kaspad/network/netadapter/router"
	"math/rand"
)

// SendAddressesContext is the interface for the context needed for the SendAddresses flow.
type SendAddressesContext interface {
	AddressManager() *addressmanager.AddressManager
}

// SendAddresses sends addresses to a peer that requests it.
func SendAddresses(context SendAddressesContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return err
	}

	msgGetAddresses := message.(*appmessage.MsgRequestAddresses)
	addresses := context.AddressManager().AddressCache(msgGetAddresses.IncludeAllSubnetworks,
		msgGetAddresses.SubnetworkID)
	msgAddresses := appmessage.NewMsgAddresses(msgGetAddresses.IncludeAllSubnetworks, msgGetAddresses.SubnetworkID)
	err = msgAddresses.AddAddresses(shuffleAddresses(addresses)...)
	if err != nil {
		return err
	}

	return outgoingRoute.Enqueue(msgAddresses)
}

// shuffleAddresses randomizes the given addresses sent if there are more than the maximum allowed in one message.
func shuffleAddresses(addresses []*appmessage.NetAddress) []*appmessage.NetAddress {
	addressCount := len(addresses)

	if addressCount < appmessage.MaxAddressesPerMsg {
		return addresses
	}

	shuffleAddresses := make([]*appmessage.NetAddress, addressCount)
	copy(shuffleAddresses, addresses)

	rand.Shuffle(addressCount, func(i, j int) {
		shuffleAddresses[i], shuffleAddresses[j] = shuffleAddresses[j], shuffleAddresses[i]
	})

	// Truncate it to the maximum size.
	shuffleAddresses = shuffleAddresses[:appmessage.MaxAddressesPerMsg]
	return shuffleAddresses
}
