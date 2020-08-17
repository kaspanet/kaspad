package protocol

import (
	"fmt"

	"github.com/kaspanet/kaspad/app/protocol/flowcontext"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/util"
)

// Manager manages the p2p protocol
type Manager struct {
	context *flowcontext.FlowContext
}

// NewManager creates a new instance of the p2p protocol manager
func NewManager(cfg *config.Config, dag *blockdag.BlockDAG, netAdapter *netadapter.NetAdapter,
	addressManager *addressmanager.AddressManager, txPool *mempool.TxPool,
	connectionManager *connmanager.ConnectionManager) (*Manager, error) {

	manager := Manager{
		context: flowcontext.New(cfg, dag, addressManager, txPool, netAdapter, connectionManager),
	}
	netAdapter.SetRouterInitializer(manager.routerInitializer)
	return &manager, nil
}

// Start starts the p2p protocol
func (m *Manager) Start() error {
	return m.context.NetAdapter().Start()
}

// Stop stops the p2p protocol
func (m *Manager) Stop() error {
	return m.context.NetAdapter().Stop()
}

// Peers returns the currently active peers
func (m *Manager) Peers() []*peerpkg.Peer {
	return m.context.Peers()
}

// IBDPeer returns the currently active IBD peer.
// Returns nil if we aren't currently in IBD
func (m *Manager) IBDPeer() *peerpkg.Peer {
	return m.context.IBDPeer()
}

// AddTransaction adds transaction to the mempool and propagates it.
func (m *Manager) AddTransaction(tx *util.Tx) error {
	return m.context.AddTransaction(tx)
}

// AddBlock adds the given block to the DAG and propagates it.
func (m *Manager) AddBlock(block *util.Block, flags blockdag.BehaviorFlags) error {
	return m.context.AddBlock(block, flags)
}

func (m *Manager) runFlows(flows []*flow, peer *peerpkg.Peer, errChan <-chan error) error {
	for _, flow := range flows {
		executeFunc := flow.executeFunc // extract to new variable so that it's not overwritten
		spawn(fmt.Sprintf("flow-%s", flow.name), func() {
			executeFunc(peer)
		})
	}

	return <-errChan
}
