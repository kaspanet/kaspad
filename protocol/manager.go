package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"time"
)

// Manager manages the p2p protocol
type Manager struct {
	netAdapter       *netadapter.NetAdapter
	txPool           *mempool.TxPool
	addedTransaction []*util.Tx
	dag              *blockdag.BlockDAG
	addressManager   *addrmgr.AddrManager

	transactionsToRebroadcast map[daghash.TxID]*util.Tx
	lastRebroadcastTime       time.Time
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(listeningAddresses []string, dag *blockdag.BlockDAG,
	addressManager *addrmgr.AddrManager, txPool *mempool.TxPool) (*Manager, error) {

	netAdapter, err := netadapter.NewNetAdapter(listeningAddresses)
	if err != nil {
		return nil, err
	}

	manager := Manager{
		netAdapter:     netAdapter,
		dag:            dag,
		addressManager: addressManager,
		txPool:         txPool,
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
