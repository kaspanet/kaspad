package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol/handlerelayblockrequests"
	"github.com/kaspanet/kaspad/protocol/handlerelayinvs"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
	"sync"
	"sync/atomic"
)

// Manager manages the p2p protocol
type Manager struct {
	netAdapter *netadapter.NetAdapter
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(listeningAddrs []string, dag *blockdag.BlockDAG) (*Manager, error) {
	netAdapter, err := netadapter.NewNetAdapter(listeningAddrs)
	if err != nil {
		return nil, err
	}

	routerInitializer := newRouterInitializer(netAdapter, dag)
	netAdapter.SetRouterInitializer(routerInitializer)

	manager := Manager{
		netAdapter: netAdapter,
	}
	return &manager, nil
}

// Start starts the p2p protocol
func (p *Manager) Start() error {
	return p.netAdapter.Start()
}

// Stop stops the p2p protocol
func (p *Manager) Stop() error {
	return p.netAdapter.Stop()
}

func newRouterInitializer(netAdapter *netadapter.NetAdapter, dag *blockdag.BlockDAG) netadapter.RouterInitializer {
	return func() (*netadapter.Router, error) {
		router := netadapter.NewRouter()
		spawn(func() {
			err := startFlows(netAdapter, router, dag)
			if err != nil {
				// TODO(libp2p) Ban peer
			}
		})
		return router, nil
	}
}

func startFlows(netAdapter *netadapter.NetAdapter, router *netadapter.Router, dag *blockdag.BlockDAG) error {
	stop := make(chan error)
	stopped := uint32(0)

	peer := new(peerpkg.Peer)

	closed, err := handshake()
	if err != nil {
		return err
	}
	if closed {
		return nil
	}

	addFlow("HandleRelayInvs", router, []string{wire.CmdInvRelayBlock, wire.CmdBlock}, &stopped, stop,
		func(ch chan wire.Message) error {
			return handlerelayinvs.HandleRelayInvs(ch, peer, netAdapter, router, dag)
		},
	)

	addFlow("HandleRelayBlockRequests", router, []string{wire.CmdGetRelayBlocks}, &stopped, stop,
		func(ch chan wire.Message) error {
			return handlerelayblockrequests.HandleRelayBlockRequests(ch, peer, router, dag)
		},
	)

	// TODO(libp2p): Remove this and change it with a real Ping-Pong flow.
	addFlow("PingPong", router, []string{wire.CmdPing, wire.CmdPong}, &stopped, stop,
		func(ch chan wire.Message) error {
			router.WriteOutgoingMessage(wire.NewMsgPing(666))
			for message := range ch {
				log.Infof("Got message: %+v", message.Command())
				if message.Command() == "ping" {
					router.WriteOutgoingMessage(wire.NewMsgPong(666))
				}
			}
			return nil
		},
	)

	err := <-stop
	return err
}

func addFlow(name string, router *netadapter.Router, messageTypes []string, stopped *uint32,
	stopChan chan error, flow func(ch chan wire.Message) error) {

	ch := make(chan wire.Message)
	err := router.AddRoute(messageTypes, ch)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		err := flow(ch)
		if err != nil {
			log.Errorf("error from %s flow: %s", name, err)
		}
		if atomic.AddUint32(stopped, 1) == 1 {
			stopChan <- err
		}
	})
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

