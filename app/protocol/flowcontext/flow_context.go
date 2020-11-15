package flowcontext

import (
	"sync"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/domain"

	"github.com/kaspanet/kaspad/app/protocol/flows/blockrelay"
	"github.com/kaspanet/kaspad/app/protocol/flows/relaytransactions"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
)

// OnBlockAddedToDAGHandler is a handler function that's triggered
// when a block is added to the DAG
type OnBlockAddedToDAGHandler func(block *externalapi.DomainBlock) error

// OnTransactionAddedToMempoolHandler is a handler function that's triggered
// when a transaction is added to the mempool
type OnTransactionAddedToMempoolHandler func(transaction *externalapi.DomainTransaction) error

// FlowContext holds state that is relevant to more than one flow or one peer, and allows communication between
// different flows that can be associated to different peers.
type FlowContext struct {
	cfg               *config.Config
	netAdapter        *netadapter.NetAdapter
	domain            domain.Domain
	addressManager    *addressmanager.AddressManager
	connectionManager *connmanager.ConnectionManager

	onBlockAddedToDAGHandler           OnBlockAddedToDAGHandler
	onTransactionAddedToMempoolHandler OnTransactionAddedToMempoolHandler

	transactionsToRebroadcastLock sync.Mutex
	transactionsToRebroadcast     map[externalapi.DomainTransactionID]*externalapi.DomainTransaction
	lastRebroadcastTime           time.Time
	sharedRequestedTransactions   *relaytransactions.SharedRequestedTransactions

	sharedRequestedBlocks *blockrelay.SharedRequestedBlocks

	isInIBD       uint32
	startIBDMutex sync.Mutex
	ibdPeer       *peerpkg.Peer

	peers      map[id.ID]*peerpkg.Peer
	peersMutex sync.RWMutex

	orphans      map[externalapi.DomainHash]*externalapi.DomainBlock
	orphansMutex sync.RWMutex
}

// New returns a new instance of FlowContext.
func New(cfg *config.Config, domain domain.Domain, addressManager *addressmanager.AddressManager,
	netAdapter *netadapter.NetAdapter, connectionManager *connmanager.ConnectionManager) *FlowContext {

	return &FlowContext{
		cfg:                         cfg,
		netAdapter:                  netAdapter,
		domain:                      domain,
		addressManager:              addressManager,
		connectionManager:           connectionManager,
		sharedRequestedTransactions: relaytransactions.NewSharedRequestedTransactions(),
		sharedRequestedBlocks:       blockrelay.NewSharedRequestedBlocks(),
		peers:                       make(map[id.ID]*peerpkg.Peer),
		transactionsToRebroadcast:   make(map[externalapi.DomainTransactionID]*externalapi.DomainTransaction),
		orphans:                     make(map[externalapi.DomainHash]*externalapi.DomainBlock),
	}
}

// SetOnBlockAddedToDAGHandler sets the onBlockAddedToDAG handler
func (f *FlowContext) SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler OnBlockAddedToDAGHandler) {
	f.onBlockAddedToDAGHandler = onBlockAddedToDAGHandler
}

// SetOnTransactionAddedToMempoolHandler sets the onTransactionAddedToMempool handler
func (f *FlowContext) SetOnTransactionAddedToMempoolHandler(onTransactionAddedToMempoolHandler OnTransactionAddedToMempoolHandler) {
	f.onTransactionAddedToMempoolHandler = onTransactionAddedToMempoolHandler
}
