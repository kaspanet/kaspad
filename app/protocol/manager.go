package protocol

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

// Manager manages the p2p protocol
type Manager struct {
	context          *flowcontext.FlowContext
	routersWaitGroup sync.WaitGroup
	isClosed         uint32
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(cfg *config.Config, domain domain.Domain, netAdapter *netadapter.NetAdapter, addressManager *addressmanager.AddressManager,
	connectionManager *connmanager.ConnectionManager) (*Manager, error) {

	manager := Manager{
		context: flowcontext.New(cfg, domain, addressManager, netAdapter, connectionManager),
	}

	netAdapter.SetP2PRouterInitializer(manager.routerInitializer)
	return &manager, nil
}

// Close closes the protocol manager and waits until all p2p flows
// finish.
func (m *Manager) Close() {
	if !atomic.CompareAndSwapUint32(&m.isClosed, 0, 1) {
		panic(errors.New("The protocol manager was already closed"))
	}

	atomic.StoreUint32(&m.isClosed, 1)
	m.context.Close()
	m.routersWaitGroup.Wait()
}

// Peers returns the currently active peers
func (m *Manager) Peers() []*peerpkg.Peer {
	return m.context.Peers()
}

// IBDPeer returns the current IBD peer or null if the node is not
// in IBD
func (m *Manager) IBDPeer() *peerpkg.Peer {
	return m.context.IBDPeer()
}

// AddTransaction adds transaction to the mempool and propagates it.
func (m *Manager) AddTransaction(tx *externalapi.DomainTransaction, allowOrphan bool) error {
	return m.context.AddTransaction(tx, allowOrphan)
}

// AddBlock adds the given block to the DAG and propagates it.
func (m *Manager) AddBlock(block *externalapi.DomainBlock) error {
	return m.context.AddBlock(block)
}

// Context returns the manager's flow context
func (m *Manager) Context() *flowcontext.FlowContext {
	return m.context
}

func (m *Manager) runFlows(flows []*common.Flow, peer *peerpkg.Peer, errChan <-chan error, flowsWaitGroup *sync.WaitGroup) error {
	flowsWaitGroup.Add(len(flows))
	for _, flow := range flows {
		executeFunc := flow.ExecuteFunc // extract to new variable so that it's not overwritten
		spawn(fmt.Sprintf("flow-%s", flow.Name), func() {
			executeFunc(peer)
			flowsWaitGroup.Done()
		})
	}

	return <-errChan
}

// SetOnVirtualChange sets the onVirtualChangeHandler handler
func (m *Manager) SetOnVirtualChange(onVirtualChangeHandler flowcontext.OnVirtualChangeHandler) {
	m.context.SetOnVirtualChangeHandler(onVirtualChangeHandler)
}

// SetOnBlockAddedToDAGHandler sets the onBlockAddedToDAG handler
func (m *Manager) SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler flowcontext.OnBlockAddedToDAGHandler) {
	m.context.SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler)
}

// SetOnNewBlockTemplateHandler sets the onNewBlockTemplate handler
func (m *Manager) SetOnNewBlockTemplateHandler(onNewBlockTemplateHandler flowcontext.OnNewBlockTemplateHandler) {
	m.context.SetOnNewBlockTemplateHandler(onNewBlockTemplateHandler)
}

// SetOnPruningPointUTXOSetOverrideHandler sets the OnPruningPointUTXOSetOverride handler
func (m *Manager) SetOnPruningPointUTXOSetOverrideHandler(onPruningPointUTXOSetOverrideHandler flowcontext.OnPruningPointUTXOSetOverrideHandler) {
	m.context.SetOnPruningPointUTXOSetOverrideHandler(onPruningPointUTXOSetOverrideHandler)
}

// SetOnTransactionAddedToMempoolHandler sets the onTransactionAddedToMempool handler
func (m *Manager) SetOnTransactionAddedToMempoolHandler(onTransactionAddedToMempoolHandler flowcontext.OnTransactionAddedToMempoolHandler) {
	m.context.SetOnTransactionAddedToMempoolHandler(onTransactionAddedToMempoolHandler)
}

// ShouldMine returns whether it's ok to use block template from this node
// for mining purposes.
func (m *Manager) ShouldMine() (bool, error) {
	return m.context.ShouldMine()
}

// IsIBDRunning returns true if IBD is currently marked as running
func (m *Manager) IsIBDRunning() bool {
	return m.context.IsIBDRunning()
}
