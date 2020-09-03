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

type Manager struct {
	context *rpccontext.Context
}

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

func (m *Manager) NotifyBlockAddedToDAG(block *util.Block) {
	m.context.BlockTemplateState.NotifyBlockAdded(block)

	notification := appmessage.NewBlockAddedNotificationMessage(block.MsgBlock())
	m.context.NotificationManager.NotifyBlockAdded(notification)
}

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
	m.context.NotificationManager.NotifyChainChanged(notification)
	return nil
}

func (m *Manager) NotifyTransactionAddedToMempool() {
	m.context.BlockTemplateState.NotifyMempoolTx()
}
