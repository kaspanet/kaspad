package handshake

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/netadapter"
	"sync"
	"sync/atomic"

	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

type Context interface {
	Config() *config.Config
	NetAdapter() *netadapter.NetAdapter
	DAG() *blockdag.BlockDAG
	AddressManager() *addrmgr.AddrManager
	StartIBDIfRequired()
}

// HandleHandshake sets up the handshake protocol - It sends a version message and waits for an incoming
// version message, as well as a verack for the sent version
func HandleHandshake(context Context, router *routerpkg.Router) (peer *peerpkg.Peer, closed bool, err error) {

	receiveVersionRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVersion})
	if err != nil {
		panic(err)
	}

	sendVersionRoute, err := router.AddIncomingRoute([]wire.MessageCommand{wire.CmdVerAck})
	if err != nil {
		panic(err)
	}

	// For HandleHandshake to finish, we need to get from the other node
	// a version and verack messages, so we increase the wait group by 2
	// and block HandleHandshake with wg.Wait().
	wg := sync.WaitGroup{}
	wg.Add(2)

	errChanUsed := uint32(0)
	errChan := make(chan error)

	peer = peerpkg.New()

	var peerAddress *wire.NetAddress
	spawn("HandleHandshake-ReceiveVersion", func() {
		defer wg.Done()
		address, err := ReceiveVersion(context, receiveVersionRoute, router.OutgoingRoute(), peer)
		if err != nil {
			log.Errorf("error from ReceiveVersion: %s", err)
		}
		if err != nil {
			if atomic.AddUint32(&errChanUsed, 1) != 1 {
				errChan <- err
			}
			return
		}
		peerAddress = address
	})

	spawn("HandleHandshake-SendVersion", func() {
		defer wg.Done()
		err := SendVersion(context, sendVersionRoute, router.OutgoingRoute())
		if err != nil {
			log.Errorf("error from SendVersion: %s", err)
		}
		if err != nil {
			if atomic.AddUint32(&errChanUsed, 1) != 1 {
				errChan <- err
			}
			return
		}
	})

	select {
	case err := <-errChan:
		if err != nil {
			return nil, false, err
		}
		return nil, true, nil
	case <-locks.ReceiveFromChanWhenDone(func() { wg.Wait() }):
	}

	err = peerpkg.AddToReadyPeers(peer)
	if err != nil {
		if errors.Is(err, peerpkg.ErrPeerWithSameIDExists) {
			return nil, false, err
		}
		panic(err)
	}

	peerID := peer.ID()
	err = context.NetAdapter().AssociateRouterID(router, peerID)
	if err != nil {
		panic(err)
	}

	if peerAddress != nil {
		subnetworkID := peer.SubnetworkID()
		context.AddressManager().AddAddress(peerAddress, peerAddress, subnetworkID)
		context.AddressManager().Good(peerAddress, subnetworkID)
	}

	context.StartIBDIfRequired()

	err = router.RemoveRoute([]wire.MessageCommand{wire.CmdVersion, wire.CmdVerAck})
	if err != nil {
		panic(err)
	}

	return peer, false, nil
}
