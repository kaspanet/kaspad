package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/receiveversion"
	"github.com/kaspanet/kaspad/protocol/sendversion"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"sync"
	"sync/atomic"
)

func handshake(router *netadapter.Router, netAdapter *netadapter.NetAdapter, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG, addressManager *addrmgr.AddrManager) (closed bool, err error) {

	receiveVersionCh := make(chan wire.Message)
	err = router.AddRoute([]string{wire.CmdVersion}, receiveVersionCh)
	if err != nil {
		panic(err)
	}
	sendVersionCh := make(chan wire.Message)

	err = router.AddRoute([]string{wire.CmdVerAck}, sendVersionCh)
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
		addr, closed, err := receiveversion.ReceiveVersion(receiveVersionCh, router, peer, dag)
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
		err := sendversion.SendVersion(sendVersionCh, router, netAdapter, dag)
		if err != nil {
			log.Errorf("error from ReceiveVersion: %s", err)
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
