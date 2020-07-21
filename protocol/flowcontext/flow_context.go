package flowcontext

import (
	"github.com/kaspanet/kaspad/addrmgr"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/netadapter/id"
	"github.com/kaspanet/kaspad/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"sync"
	"time"
)

// FlowContext holds state that is relevant to more than one flow or one peer, and allows communication between
// different flows that can be associated to different peers.
type FlowContext struct {
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

	isInIBD       uint32
	startIBDMutex sync.Mutex

	readyPeers      map[*id.ID]*peerpkg.Peer
	readyPeersMutex sync.RWMutex
}

// New returns a new instance of FlowContext.
func New(cfg *config.Config, dag *blockdag.BlockDAG,
	addressManager *addrmgr.AddrManager, txPool *mempool.TxPool, netAdapter *netadapter.NetAdapter) *FlowContext {
	return &FlowContext{
		cfg:                         cfg,
		netAdapter:                  netAdapter,
		dag:                         dag,
		addressManager:              addressManager,
		txPool:                      txPool,
		sharedRequestedTransactions: relaytransactions.NewSharedRequestedTransactions(),
		sharedRequestedBlocks:       blockrelay.NewSharedRequestedBlocks(),
	}
}
