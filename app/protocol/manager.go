package protocol

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"

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
	context   *flowcontext.FlowContext
	routersWg sync.WaitGroup
	isClosed  uint32
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
		panic(errors.New("The logger is already running"))
	}

	atomic.StoreUint32(&m.isClosed, 1)
	m.context.Close()
	m.routersWg.Wait()
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
func (m *Manager) AddTransaction(tx *externalapi.DomainTransaction) error {
	return m.context.AddTransaction(tx)
}

// AddBlock adds the given block to the DAG and propagates it.
func (m *Manager) AddBlock(block *externalapi.DomainBlock) error {
	return m.context.AddBlock(block)
}

func (m *Manager) runFlows(flows []*flow, peer *peerpkg.Peer, errChan <-chan error, flowsWaitGroup *sync.WaitGroup) error {
	flowsWaitGroup.Add(len(flows))
	for _, flow := range flows {
		executeFunc := flow.executeFunc // extract to new variable so that it's not overwritten
		spawn(fmt.Sprintf("flow-%s", flow.name), func() {
			executeFunc(peer)
			flowsWaitGroup.Done()
		})
	}

	return <-errChan
}

// SetOnBlockAddedToDAGHandler sets the onBlockAddedToDAG handler
func (m *Manager) SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler flowcontext.OnBlockAddedToDAGHandler) {
	m.context.SetOnBlockAddedToDAGHandler(onBlockAddedToDAGHandler)
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
