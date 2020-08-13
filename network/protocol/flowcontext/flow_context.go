package flowcontext

import (
	"sync"
	"time"

	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/network/addressmanager"
	"github.com/kaspanet/kaspad/network/connmanager"
	"github.com/kaspanet/kaspad/network/netadapter"
	"github.com/kaspanet/kaspad/network/netadapter/id"
	"github.com/kaspanet/kaspad/network/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/network/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/network/protocol/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// FlowContext holds state that is relevant to more than one flow or one peer, and allows communication between
// different flows that can be associated to different peers.
type FlowContext struct {
	cfg               *config.Config
	netAdapter        *netadapter.NetAdapter
	txPool            *mempool.TxPool
	dag               *blockdag.BlockDAG
	addressManager    *addressmanager.AddressManager
	connectionManager *connmanager.ConnectionManager

	transactionsToRebroadcastLock sync.Mutex
	transactionsToRebroadcast     map[daghash.TxID]*util.Tx
	lastRebroadcastTime           time.Time
	sharedRequestedTransactions   *relaytransactions.SharedRequestedTransactions

	sharedRequestedBlocks *blockrelay.SharedRequestedBlocks

	isInIBD       uint32
	startIBDMutex sync.Mutex
	ibdPeer       *peerpkg.Peer

	peers      map[*id.ID]*peerpkg.Peer
	peersMutex sync.RWMutex
}

// New returns a new instance of FlowContext.
func New(cfg *config.Config, dag *blockdag.BlockDAG, addressManager *addressmanager.AddressManager,
	txPool *mempool.TxPool, netAdapter *netadapter.NetAdapter,
	connectionManager *connmanager.ConnectionManager) *FlowContext {

	return &FlowContext{
		cfg:                         cfg,
		netAdapter:                  netAdapter,
		dag:                         dag,
		addressManager:              addressManager,
		connectionManager:           connectionManager,
		txPool:                      txPool,
		sharedRequestedTransactions: relaytransactions.NewSharedRequestedTransactions(),
		sharedRequestedBlocks:       blockrelay.NewSharedRequestedBlocks(),
		peers:                       make(map[*id.ID]*peerpkg.Peer),
		transactionsToRebroadcast:   make(map[daghash.TxID]*util.Tx),
	}
}
