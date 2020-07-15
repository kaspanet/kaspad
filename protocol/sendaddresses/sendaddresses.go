package sendaddresses

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
	"math/rand"
	"time"
)

// SendAddresses sends addresses to a peer that requests it.
func SendAddresses(incomingRoute *router.Route, outgoingRoute *router.Route,
	addressManager *addrmgr.AddrManager) (routeClosed bool, err error) {

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return true, nil
	}

	msgGetAddresses := message.(*wire.MsgGetAddresses)
	addresses := addressManager.AddressCache(msgGetAddresses.IncludeAllSubnetworks, msgGetAddresses.SubnetworkID)
	msgAddr := wire.NewMsgAddr(msgGetAddresses.IncludeAllSubnetworks, msgGetAddresses.SubnetworkID)
	err = msgAddr.AddAddresses(shuffleAddresses(addresses)...)
	if err != nil {
		panic(err)
	}

	const timeout = 30 * time.Second
	isOpen, err = outgoingRoute.EnqueueWithTimeout(msgAddr, timeout)
	if err != nil {
		return false, err
	}
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
