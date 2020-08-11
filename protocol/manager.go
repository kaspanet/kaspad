package protocol

import (
	"fmt"

	"github.com/kaspanet/kaspad/addressmanager"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/config"
	"github.com/kaspanet/kaspad/connmanager"
	"github.com/kaspanet/kaspad/mempool"
	"github.com/kaspanet/kaspad/netadapter"
	"github.com/kaspanet/kaspad/protocol/flowcontext"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
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
	err := m.context.AddTransaction(tx)
	if err != nil {
		if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
			return err
		}

		panic(err)
	}
	return nil
}

// AddBlock adds the given block to the DAG and propagates it.
func (m *Manager) AddBlock(block *util.Block, flags blockdag.BehaviorFlags) error {
	err := m.context.AddBlock(block, flags)
	if err != nil {
		if protocolErr := &(protocolerrors.ProtocolError{}); errors.As(err, &protocolErr) {
			return err
		}

		panic(err)
	}
	return nil
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
