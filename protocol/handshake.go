package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	routerpkg "github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/receiveversion"
	"github.com/kaspanet/kaspad/protocol/sendversion"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
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

	wg := sync.WaitGroup{}
	wg.Add(2)

	var (
		errChanUsed uint32
		errChan     = make(chan error)
	)

	var peerAddr *wire.NetAddress
	spawn(func() {
		defer wg.Done()
		addr, closed, err := receiveversion.ReceiveVersion(receiveVersionRoute, router.OutgoingRoute(), peer, dag)
		if err != nil || closed {
			if err != nil {
				log.Errorf("error from ReceiveVersion: %s", err)
			}
			if atomic.AddUint32(&errChanUsed, 1) != 1 {
				errChan <- err
			}
			return
		}
		peerAddr = addr
	})

	spawn(func() {
		defer wg.Done()
		closed, err := sendversion.SendVersion(sendVersionRoute, router.OutgoingRoute(), netAdapter, dag)
		if err != nil || closed {
			log.Errorf("error from SendVersion: %s", err)
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
	case <-locks.TickWhenDone(func() { wg.Wait() }):
	}

	err = peer.MarkAsReady()
	if err != nil {
		panic(err)
	}

	if peerAddr != nil {
		subnetworkID, err := peer.SubnetworkID()
		if err != nil {
			panic(err)
		}
		addressManager.AddAddress(peerAddr, peerAddr, subnetworkID)
	}
	return false, nil
}
