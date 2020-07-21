package protocol

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
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

	// TODO(libp2p) populate these vars
	isInIBD uint32
	ibdPeer *peerpkg.Peer

	peers *peerpkg.Peers
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
		peers:                       peerpkg.NewPeers(),
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

// Peers returns the currently active peers
func (m *Manager) Peers() []*peerpkg.Peer {
	return m.peers.ReadyPeers()
}

// IBDPeer returns the currently active IBD peer.
// Returns nil if we aren't currently in IBD
func (m *Manager) IBDPeer() *peerpkg.Peer {
	return m.ibdPeer
}
