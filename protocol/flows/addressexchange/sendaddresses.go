package addressexchange

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
	"math/rand"
)

type SendAddressesContext interface {
	AddressManager() *addrmgr.AddrManager
}

// SendAddresses sends addresses to a peer that requests it.
func SendAddresses(context SendAddressesContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return err
	}

	msgGetAddresses := message.(*wire.MsgGetAddresses)
	addresses := context.AddressManager().AddressCache(msgGetAddresses.IncludeAllSubnetworks,
		msgGetAddresses.SubnetworkID)
	msgAddresses := wire.NewMsgAddresses(msgGetAddresses.IncludeAllSubnetworks, msgGetAddresses.SubnetworkID)
	err = msgAddresses.AddAddresses(shuffleAddresses(addresses)...)
	if err != nil {
		panic(err)
	}

	return outgoingRoute.Enqueue(msgAddresses)
}

// shuffleAddresses randomizes the given addresses sent if there are more than the maximum allowed in one message.
func shuffleAddresses(addresses []*wire.NetAddress) []*wire.NetAddress {
	addressCount := len(addresses)

	if addressCount < wire.MaxAddressesPerMsg {
		return addresses
	}

	shuffleAddresses := make([]*wire.NetAddress, addressCount)
	copy(shuffleAddresses, addresses)

	rand.Shuffle(addressCount, func(i, j int) {
		shuffleAddresses[i], shuffleAddresses[j] = shuffleAddresses[j], shuffleAddresses[i]
	})

	// Truncate it to the maximum size.
	shuffleAddresses = shuffleAddresses[:wire.MaxAddressesPerMsg]
	return shuffleAddresses
}
