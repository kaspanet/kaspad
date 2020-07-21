package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"sync"
	"time"
)

// Manager manages the p2p protocol
type Manager struct {
	cfg               *config.Config
	netAdapter        *netadapter.NetAdapter
	txPool            *mempool.TxPool
	addedTransactions []*util.Tx
	dag               *blockdag.BlockDAG
	addressManager    *addrmgr.AddrManager

	transactionsToRebroadcastLock sync.Mutex
	transactionsToRebroadcast     map[daghash.TxID]*util.Tx
	lastRebroadcastTime           time.Time
	sharedRequestedTransactions   *relaytransactions.SharedRequestedTransactions

	sharedRequestedBlocks *blockrelay.SharedRequestedBlocks

	isInIBD uint32 // TODO(libp2p) populate this var
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(cfg *config.Config, dag *blockdag.BlockDAG,
	addressManager *addrmgr.AddrManager, txPool *mempool.TxPool) (*Manager, error) {

	netAdapter, err := netadapter.NewNetAdapter(cfg)
	if err != nil {
		return nil, err
	}

	manager := Manager{
		netAdapter:                  netAdapter,
		dag:                         dag,
		addressManager:              addressManager,
		txPool:                      txPool,
		sharedRequestedTransactions: relaytransactions.NewSharedRequestedTransactions(),
		sharedRequestedBlocks:       blockrelay.NewSharedRequestedBlocks(),
	}
	netAdapter.SetRouterInitializer(manager.routerInitializer)
	return &manager, nil
}

// Start starts the p2p protocol
func (m *Manager) Start() error {
	return m.netAdapter.Start()
}

// Stop stops the p2p protocol
func (m *Manager) Stop() error {
	return m.netAdapter.Stop()
}

func (m *Manager) Config() *config.Config {
	return m.cfg
}

func (m *Manager) NetAdapter() *netadapter.NetAdapter {
	return m.netAdapter
}
