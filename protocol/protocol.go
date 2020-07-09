package protocol

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter"
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
		stop := make(chan struct{})
		shutdown := uint32(0)

		blockRelayCh := make(chan wire.Message)
		mux.AddFlow([]string{wire.CmdInvRelayBlock, wire.CmdBlock}, blockRelayCh)
		spawn(func() {
			err := blockrelay.StartBlockRelay(blockRelayCh, server, connection, dag)
			if err == nil {
				return
			}

			log.Errorf("error from StartBlockRelay: %s", err)
			if atomic.LoadUint32(&shutdown) == 0 {
				stop <- struct{}{}
			}
		})

		getRelayBlocksListenerCh := make(chan wire.Message)
		mux.AddFlow([]string{wire.CmdGetRelayBlocks}, getRelayBlocksListenerCh)
		spawn(func() {
			err := getrelayblockslistener.StartGetRelayBlocksListener(getRelayBlocksListenerCh, connection, dag)
			if err == nil {
				return
			}

			log.Errorf("error from StartGetRelayBlocksListener: %s", err)
			if atomic.LoadUint32(&shutdown) == 0 {
				stop <- struct{}{}
			}
		})

		<-stop
		atomic.StoreUint32(&shutdown, 1)
		return router, nil
	}
}
