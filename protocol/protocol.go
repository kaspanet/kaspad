package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/messagemux"
	"github.com/kaspanet/kaspad/p2pserver"
	"github.com/kaspanet/kaspad/protocol/blockrelay"
	"github.com/kaspanet/kaspad/protocol/getrelayblockslistener"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/receiveversion"
	"github.com/kaspanet/kaspad/protocol/sendversion"
	"github.com/kaspanet/kaspad/util/locks"
	"github.com/kaspanet/kaspad/wire"
	"sync"
	"sync/atomic"
)

// StartProtocol starts the p2p protocol for a given connection
func StartProtocol(server p2pserver.Server, mux messagemux.Mux, connection p2pserver.Connection,
	dag *blockdag.BlockDAG) error {

	closed, err := handshake()
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	errChan := make(chan error)
	stopped := uint32(0)

	blockRelayCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdInvRelayBlock, wire.CmdBlock}, blockRelayCh)
	spawn(func() {
		err := blockrelay.StartBlockRelay(blockRelayCh, server, connection, dag)
		if err != nil {
			log.Errorf("error from StartBlockRelay: %s", err)
		}

		if atomic.AddUint32(&stopped, 1) != 1 {
			errChan <- err
		}
	})

	getRelayBlocksListenerCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdGetRelayBlocks}, getRelayBlocksListenerCh)
	spawn(func() {
		err := getrelayblockslistener.StartGetRelayBlocksListener(getRelayBlocksListenerCh, connection, dag)
		if err != nil {
			log.Errorf("error from StartGetRelayBlocksListener: %s", err)
		}

		if atomic.AddUint32(&stopped, 1) != 1 {
			errChan <- err
		}
	})

	err = <-errChan
	return err
}

func handshake(server p2pserver.Server, mux messagemux.Mux, connection p2pserver.Connection, peer *peerpkg.Peer,
	dag *blockdag.BlockDAG, addressManager *addrmgr.AddrManager) (closed bool, err error) {

	receiveVersionCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdVersion}, receiveVersionCh)
	sendVersionCh := make(chan wire.Message)
	mux.AddFlow([]string{wire.CmdVerAck}, sendVersionCh)

	wg := sync.WaitGroup{}
	wg.Add(2)

	var (
		errChanUsed uint32
		errChan     = make(chan error)
	)

	var peerAddr *wire.NetAddress
	spawn(func() {
		defer wg.Done()
		addr, closed, err := receiveversion.ReceiveVersion(receiveVersionCh, connection, peer, dag)
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
		err := sendversion.SendVersion(sendVersionCh, connection, peer, dag)
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
