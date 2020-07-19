package handshake

import (
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/ibd"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// HandleHandshake sets up the handshake protocol - It sends a version message and waits for an incoming
// version message, as well as a verack for the sent version
func HandleHandshake(router *routerpkg.Router, netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG, addressManager *addrmgr.AddrManager) (peer *peerpkg.Peer, closed bool, err error) {

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

	var peerAddress *wire.NetAddress
	spawn(func() {
		defer wg.Done()
		address, closed, err := ReceiveVersion(receiveVersionRoute, router.OutgoingRoute(), netAdapter, peer, dag)
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
		closed, err := SendVersion(sendVersionRoute, router.OutgoingRoute(), netAdapter, dag)
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
			return nil, false, err
		}
		return nil, true, nil
	case <-locks.ReceiveFromChanWhenDone(func() { wg.Wait() }):
	}

	peer = peerpkg.New()
	err = peerpkg.AddToReadyPeers(peer)
	if err != nil {
		if errors.Is(err, peerpkg.ErrPeerWithSameIDExists) {
			return nil, false, err
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
		addressManager.Good(peerAddress, subnetworkID)
	}

	err = ibd.StartIBDIfRequired(dag)
	if err != nil {
		return nil, false, err
	}

	err = router.RemoveRoute([]wire.MessageCommand{wire.CmdVersion, wire.CmdVerAck})
	if err != nil {
		panic(err)
	}

	return peer, false, nil
}
