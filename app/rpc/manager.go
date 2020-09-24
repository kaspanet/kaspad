package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/blockdag/indexers"
	"github.com/kaspanet/kaspad/domain/mempool"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// Manager is an RPC manager
type Manager struct {
	context *rpccontext.Context
}

// NewManager creates a new RPC Manager
func NewManager(
	cfg *config.Config,
	netAdapter *netadapter.NetAdapter,
	dag *blockdag.BlockDAG,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	blockTemplateGenerator *mining.BlkTmplGenerator,
	mempool *mempool.TxPool,
	addressManager *addressmanager.AddressManager,
	acceptanceIndex *indexers.AcceptanceIndex) *Manager {

	manager := Manager{
		context: rpccontext.NewContext(
			cfg,
			netAdapter,
			dag,
			protocolManager,
			connectionManager,
			blockTemplateGenerator,
			mempool,
			addressManager,
			acceptanceIndex,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}

// NotifyBlockAddedToDAG notifies the manager that a block has been added to the DAG
func (m *Manager) NotifyBlockAddedToDAG(block *util.Block) error {
	m.context.BlockTemplateState.NotifyBlockAdded(block)

	notification := appmessage.NewBlockAddedNotificationMessage(block.MsgBlock())
	return m.context.NotificationManager.NotifyBlockAdded(notification)
}

// NotifyChainChanged notifies the manager that the DAG's selected parent chain has changed
func (m *Manager) NotifyChainChanged(removedChainBlockHashes []*daghash.Hash, addedChainBlockHashes []*daghash.Hash) error {
	addedChainBlocks, err := m.context.CollectChainBlocks(addedChainBlockHashes)
	if err != nil {
		return err
	}
	removedChainBlockHashStrings := make([]string, len(removedChainBlockHashes))
	for i, removedChainBlockHash := range removedChainBlockHashes {
		removedChainBlockHashStrings[i] = removedChainBlockHash.String()
	}
	notification := appmessage.NewChainChangedNotificationMessage(removedChainBlockHashStrings, addedChainBlocks)
	return m.context.NotificationManager.NotifyChainChanged(notification)
}

// NotifyFinalityConflict notifies the manager that there's a finality conflict in the DAG
func (m *Manager) NotifyFinalityConflict(violatingBlockHash string) error {
	notification := appmessage.NewFinalityConflictNotificationMessage(violatingBlockHash)
	return m.context.NotificationManager.NotifyFinalityConflict(notification)
}

// NotifyFinalityConflictResolved notifies the manager that a finality conflict in the DAG has been resolved
func (m *Manager) NotifyFinalityConflictResolved(finalityBlockHash string) error {
	notification := appmessage.NewFinalityConflictResolvedNotificationMessage(finalityBlockHash)
	return m.context.NotificationManager.NotifyFinalityConflictResolved(notification)
}

// NotifyTransactionAddedToMempool notifies the manager that a transaction has been added to the mempool
func (m *Manager) NotifyTransactionAddedToMempool() {
	m.context.BlockTemplateState.NotifyMempoolTx()
}
