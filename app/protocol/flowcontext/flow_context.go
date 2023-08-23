package flowcontext

import (
	"sync"
	"time"

	"github.com/c4ei/yunseokyeol/util/mstime"

	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"

	"github.com/c4ei/yunseokyeol/domain"

	peerpkg "github.com/c4ei/yunseokyeol/app/protocol/peer"
	"github.com/c4ei/yunseokyeol/infrastructure/config"
	"github.com/c4ei/yunseokyeol/infrastructure/network/addressmanager"
	"github.com/c4ei/yunseokyeol/infrastructure/network/connmanager"
	"github.com/c4ei/yunseokyeol/infrastructure/network/netadapter"
	"github.com/c4ei/yunseokyeol/infrastructure/network/netadapter/id"
)

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

// IsNearlySynced returns whether current consensus is considered synced or close to being synced.
func (f *FlowContext) IsNearlySynced() (bool, error) {
	return f.Domain().Consensus().IsNearlySynced()
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
