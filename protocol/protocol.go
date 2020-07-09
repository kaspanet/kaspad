package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol/blockrelay"
	"github.com/kaspanet/kaspad/protocol/getrelayblockslistener"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/wire"
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
			err := startFlow(netAdapter, router, dag)
			if err != nil {
				// TODO(libp2p) Ban peer
			}
		})
		return router, nil
	}
}

func startFlow(netAdapter *netadapter.NetAdapter, router *netadapter.Router, dag *blockdag.BlockDAG) error {
	stop := make(chan error)
	shutdown := uint32(0)

	peer := new(peerpkg.Peer)

	blockRelayCh := make(chan wire.Message)
	err := router.AddRoute([]string{wire.CmdInvRelayBlock, wire.CmdBlock}, blockRelayCh)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		err := blockrelay.StartBlockRelay(blockRelayCh, peer, netAdapter, router, dag)
		if err != nil {
			log.Errorf("error from StartBlockRelay: %s", err)
		}

		if atomic.AddUint32(&shutdown, 1) == 1 {
			stop <- err
		}
	})

	getRelayBlocksListenerCh := make(chan wire.Message)
	err = router.AddRoute([]string{wire.CmdGetRelayBlocks}, getRelayBlocksListenerCh)
	if err != nil {
		panic(err)
	}

	spawn(func() {
		err := getrelayblockslistener.StartGetRelayBlocksListener(getRelayBlocksListenerCh, router, dag)
		if err != nil {
			log.Errorf("error from StartGetRelayBlocksListener: %s", err)
		}

		if atomic.AddUint32(&shutdown, 1) == 1 {
			stop <- err
		}
	})

	err = <-stop
	return err
}
