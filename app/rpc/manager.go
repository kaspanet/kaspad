package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
)

// Manager is an RPC manager
type Manager struct {
	context *rpccontext.Context
}

// NewManager creates a new RPC Manager
func NewManager(
	cfg *config.Config,
	domain domain.Domain,
	netAdapter *netadapter.NetAdapter,
	protocolManager *protocol.Manager,
	connectionManager *connmanager.ConnectionManager,
	addressManager *addressmanager.AddressManager,
	shutDownChan chan<- struct{}) *Manager {

	manager := Manager{
		context: rpccontext.NewContext(
			cfg,
			domain,
			netAdapter,
			protocolManager,
			connectionManager,
			addressManager,
			shutDownChan,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}

// NotifyBlockAddedToDAG notifies the manager that a block has been added to the DAG
func (m *Manager) NotifyBlockAddedToDAG(block *externalapi.DomainBlock) error {
	notification := appmessage.NewBlockAddedNotificationMessage(appmessage.DomainBlockToMsgBlock(block))
	return m.context.NotificationManager.NotifyBlockAdded(notification)
}

// NotifyChainChanged notifies the manager that the DAG's selected parent chain has changed
func (m *Manager) NotifyChainChanged(removedChainBlockHashes []*externalapi.DomainHash,
	addedChainBlockHashes []*externalapi.DomainHash) error {
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
