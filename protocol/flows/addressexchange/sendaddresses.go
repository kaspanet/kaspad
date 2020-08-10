package addressexchange

import (
	"github.com/kaspanet/kaspad/addressmanager"
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/router"
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

	msgGetAddresses := message.(*domainmessage.MsgRequestAddresses)
	addresses := context.AddressManager().AddressCache(msgGetAddresses.IncludeAllSubnetworks,
		msgGetAddresses.SubnetworkID)
	msgAddresses := domainmessage.NewMsgAddresses(msgGetAddresses.IncludeAllSubnetworks, msgGetAddresses.SubnetworkID)
	err = msgAddresses.AddAddresses(shuffleAddresses(addresses)...)
	if err != nil {
		return err
	}

	return outgoingRoute.Enqueue(msgAddresses)
}

// shuffleAddresses randomizes the given addresses sent if there are more than the maximum allowed in one message.
func shuffleAddresses(addresses []*domainmessage.NetAddress) []*domainmessage.NetAddress {
	addressCount := len(addresses)

	if addressCount < domainmessage.MaxAddressesPerMsg {
		return addresses
	}

	shuffleAddresses := make([]*domainmessage.NetAddress, addressCount)
	copy(shuffleAddresses, addresses)

	rand.Shuffle(addressCount, func(i, j int) {
		shuffleAddresses[i], shuffleAddresses[j] = shuffleAddresses[j], shuffleAddresses[i]
	})

	// Truncate it to the maximum size.
	shuffleAddresses = shuffleAddresses[:domainmessage.MaxAddressesPerMsg]
	return shuffleAddresses
}
