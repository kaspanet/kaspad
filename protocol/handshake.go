package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/flows/receiveversion"
	"github.com/kaspanet/kaspad/protocol/flows/sendversion"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

func handshake(router *routerpkg.Router, netAdapter *netadapter.NetAdapter, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG, addressManager *addrmgr.AddrManager) (closed bool, err error) {

	receiveVersionRoute, err := router.AddIncomingRoute([]string{wire.CmdVersion})
	if err != nil {
		panic(err)
	}

	sendVersionRoute, err := router.AddIncomingRoute([]string{wire.CmdVerAck})
	if err != nil {
		panic(err)
	}

	// For the handshake to finish, we need to get from the other node
	// a version and verack messages, so we increase the wait group by 2
	// and block the handshake with wg.Wait().
	wg := sync.WaitGroup{}
	wg.Add(2)

	errChanUsed := uint32(0)
	errChan := make(chan error)

	var peerAddress *wire.NetAddress
	spawn(func() {
		defer wg.Done()
		address, closed, err := receiveversion.ReceiveVersion(receiveVersionRoute, router.OutgoingRoute(), netAdapter, peer, dag)
		if err != nil {
			log.Errorf("error from ReceiveVersion: %s", err)
		}
		if err != nil || closed {
			if atomic.AddUint32(&errChanUsed, 1) != 1 {
				errChan <- err
			}
			return
		}
		peerAddress = address
	})

	spawn(func() {
		defer wg.Done()
		closed, err := sendversion.SendVersion(sendVersionRoute, router.OutgoingRoute(), netAdapter, dag)
		if err != nil {
			log.Errorf("error from SendVersion: %s", err)
		}
		if err != nil || closed {
			if atomic.AddUint32(&errChanUsed, 1) != 1 {
				errChan <- err
			}
			return
		}
	})

	select {
	case err := <-errChan:
		if err != nil {
			return false, err
		}
		return true, nil
	case <-locks.ReceiveFromChanWhenDone(func() { wg.Wait() }):
	}

	err = peerpkg.AddToReadyPeers(peer)
	if err != nil {
		if errors.Is(err, peerpkg.ErrPeerWithSameIDExists) {
			return false, err
		}
		panic(err)
	}

	peerID, err := peer.ID()
	if err != nil {
		panic(err)
	}

	err = netAdapter.AssociateRouterID(router, peerID)
	if err != nil {
		panic(err)
	}

	if peerAddress != nil {
		subnetworkID, err := peer.SubnetworkID()
		if err != nil {
			panic(err)
		}
		addressManager.AddAddress(peerAddress, peerAddress, subnetworkID)
	}

	err = router.RemoveRoute([]string{wire.CmdVersion, wire.CmdVerAck})
	if err != nil {
		panic(err)
	}
	return false, nil
}
