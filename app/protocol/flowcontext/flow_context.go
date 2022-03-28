package flowcontext

import (
	"sync"
	"time"

	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/domain"

	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
)

// OnBlockAddedToDAGHandler is a handler function that's triggered
// when a block is added to the DAG
type OnBlockAddedToDAGHandler func(block *externalapi.DomainBlock, virtualChangeSet *externalapi.VirtualChangeSet) error

// OnVirtualChangeHandler is a handler function that's triggered when the virtual changes
type OnVirtualChangeHandler func(virtualChangeSet *externalapi.VirtualChangeSet) error

// OnNewBlockTemplateHandler is a handler function that's triggered when a new block template is available
type OnNewBlockTemplateHandler func() error

// OnPruningPointUTXOSetOverrideHandler is a handle function that's triggered whenever the UTXO set
// resets due to pruning point change via IBD.
type OnPruningPointUTXOSetOverrideHandler func() error

// OnTransactionAddedToMempoolHandler is a handler function that's triggered
// when a transaction is added to the mempool
type OnTransactionAddedToMempoolHandler func()

// FlowContext holds state that is relevant to more than one flow or one peer, and allows communication between
// different flows that can be associated to different peers.
type FlowContext struct {
	cfg               *config.Config
	netAdapter        *netadapter.NetAdapter
	domain            domain.Domain
	addressManager    *addressmanager.AddressManager
	connectionManager *connmanager.ConnectionManager

	timeStarted int64

	onVirtualChangeHandler               OnVirtualChangeHandler
	onBlockAddedToDAGHandler             OnBlockAddedToDAGHandler
	onNewBlockTemplateHandler            OnNewBlockTemplateHandler
	onPruningPointUTXOSetOverrideHandler OnPruningPointUTXOSetOverrideHandler
	onTransactionAddedToMempoolHandler   OnTransactionAddedToMempoolHandler

	lastRebroadcastTime         time.Time
	sharedRequestedTransactions *SharedRequestedTransactions

	sharedRequestedBlocks *SharedRequestedBlocks

	ibdPeer      *peerpkg.Peer
	ibdPeerMutex sync.RWMutex

	peers      map[id.ID]*peerpkg.Peer
	peersMutex sync.RWMutex

	orphans      map[externalapi.DomainHash]*externalapi.DomainBlock
	orphansMutex sync.RWMutex

	transactionIDsToPropagate        []*externalapi.DomainTransactionID
	lastTransactionIDPropagationTime time.Time
	transactionIDPropagationLock     sync.Mutex

	shutdownChan chan struct{}
}

// New returns a new instance of FlowContext.
func New(cfg *config.Config, domain domain.Domain, addressManager *addressmanager.AddressManager,
	netAdapter *netadapter.NetAdapter, connectionManager *connmanager.ConnectionManager) *FlowContext {

	return &FlowContext{
		cfg:                              cfg,
		netAdapter:                       netAdapter,
		domain:                           domain,
		addressManager:                   addressManager,
		connectionManager:                connectionManager,
		sharedRequestedTransactions:      NewSharedRequestedTransactions(),
		sharedRequestedBlocks:            NewSharedRequestedBlocks(),
		peers:                            make(map[id.ID]*peerpkg.Peer),
		orphans:                          make(map[externalapi.DomainHash]*externalapi.DomainBlock),
		timeStarted:                      mstime.Now().UnixMilliseconds(),
		transactionIDsToPropagate:        []*externalapi.DomainTransactionID{},
		lastTransactionIDPropagationTime: time.Now(),
		shutdownChan:                     make(chan struct{}),
	}
}

// Close signals to all flows the the protocol manager is closed.
func (f *FlowContext) Close() {
	close(f.shutdownChan)
}

// ShutdownChan is a chan where flows can subscribe to shutdown
// event.
func (f *FlowContext) ShutdownChan() <-chan struct{} {
	return f.shutdownChan
}

// SetOnVirtualChangeHandler sets the onVirtualChangeHandler handler
func (f *FlowContext) SetOnVirtualChangeHandler(onVirtualChangeHandler OnVirtualChangeHandler) {
	f.onVirtualChangeHandler = onVirtualChangeHandler
}

// SetOnBlockAddedToDAGHandler sets the onBlockAddedToDAG handler
func (f *FlowContext) SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler OnBlockAddedToDAGHandler) {
	f.onBlockAddedToDAGHandler = onBlockAddedToDAGHandler
}

// SetOnNewBlockTemplateHandler sets the onNewBlockTemplateHandler handler
func (f *FlowContext) SetOnNewBlockTemplateHandler(onNewBlockTemplateHandler OnNewBlockTemplateHandler) {
	f.onNewBlockTemplateHandler = onNewBlockTemplateHandler
}

// SetOnPruningPointUTXOSetOverrideHandler sets the onPruningPointUTXOSetOverrideHandler handler
func (f *FlowContext) SetOnPruningPointUTXOSetOverrideHandler(onPruningPointUTXOSetOverrideHandler OnPruningPointUTXOSetOverrideHandler) {
	f.onPruningPointUTXOSetOverrideHandler = onPruningPointUTXOSetOverrideHandler
}

// SetOnTransactionAddedToMempoolHandler sets the onTransactionAddedToMempool handler
func (f *FlowContext) SetOnTransactionAddedToMempoolHandler(onTransactionAddedToMempoolHandler OnTransactionAddedToMempoolHandler) {
	f.onTransactionAddedToMempoolHandler = onTransactionAddedToMempoolHandler
}
