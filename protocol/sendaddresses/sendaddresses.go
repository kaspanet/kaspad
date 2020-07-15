package sendaddresses

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
	"math/rand"
)

// SendAddresses sends address to a peer that requests it.
func SendAddresses(incomingRoute *router.Route, outgoingRoute *router.Route,
	addressManager *addrmgr.AddrManager) (routeClosed bool, err error) {

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return true, nil
	}

	msgGetAddr := message.(*wire.MsgGetAddr)
	addresses := addressManager.AddressCache(msgGetAddr.IncludeAllSubnetworks, msgGetAddr.SubnetworkID)
	msgAddr := wire.NewMsgAddr(msgGetAddr.IncludeAllSubnetworks, msgGetAddr.SubnetworkID)
	err = msgAddr.AddAddresses(shuffleAddresses(addresses)...)
	if err != nil {
		panic(err)
	}

	isOpen = outgoingRoute.Enqueue(msgAddr)
	if !isOpen {
		return true, nil
	}
	return false, nil
}

// shuffleAddresses randomizes the given addresses sent if there are more than the maximum allowed in one message.
func shuffleAddresses(addresses []*wire.NetAddress) []*wire.NetAddress {
	addressCount := len(addresses)

	if addressCount < wire.MaxAddrPerMsg {
		return addresses
	}

	shuffleAddresses := make([]*wire.NetAddress, addressCount)
	copy(shuffleAddresses, addresses)

	rand.Shuffle(addressCount, func(i, j int) {
		shuffleAddresses[i], shuffleAddresses[j] = shuffleAddresses[j], shuffleAddresses[i]
	})

	// Truncate it to the maximum size.
	shuffleAddresses = shuffleAddresses[:wire.MaxAddrPerMsg]
	return shuffleAddresses
}
