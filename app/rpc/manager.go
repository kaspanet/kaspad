package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/utxoindex"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/logger"
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
	utxoIndex *utxoindex.UTXOIndex,
	shutDownChan chan<- struct{}) *Manager {

	manager := Manager{
		context: rpccontext.NewContext(
			cfg,
			domain,
			netAdapter,
			protocolManager,
			connectionManager,
			addressManager,
			utxoIndex,
			shutDownChan,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	return &manager
}

// NotifyBlockAddedToDAG notifies the manager that a block has been added to the DAG
func (m *Manager) NotifyBlockAddedToDAG(block *externalapi.DomainBlock, blockInsertionResult *externalapi.BlockInsertionResult) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyBlockAddedToDAG")
	defer onEnd()

	if m.context.Config.UTXOIndex {
		err := m.notifyUTXOsChanged(blockInsertionResult)
		if err != nil {
			return err
		}
	}

	err := m.notifyVirtualSelectedParentBlueScoreChanged()
	if err != nil {
		return err
	}

	blockAddedNotification := appmessage.NewBlockAddedNotificationMessage(appmessage.DomainBlockToMsgBlock(block))
	return m.context.NotificationManager.NotifyBlockAdded(blockAddedNotification)
}

// NotifyFinalityConflict notifies the manager that there's a finality conflict in the DAG
func (m *Manager) NotifyFinalityConflict(violatingBlockHash string) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyFinalityConflict")
	defer onEnd()

	notification := appmessage.NewFinalityConflictNotificationMessage(violatingBlockHash)
	return m.context.NotificationManager.NotifyFinalityConflict(notification)
}

// NotifyFinalityConflictResolved notifies the manager that a finality conflict in the DAG has been resolved
func (m *Manager) NotifyFinalityConflictResolved(finalityBlockHash string) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyFinalityConflictResolved")
	defer onEnd()

	notification := appmessage.NewFinalityConflictResolvedNotificationMessage(finalityBlockHash)
	return m.context.NotificationManager.NotifyFinalityConflictResolved(notification)
}

func (m *Manager) notifyUTXOsChanged(blockInsertionResult *externalapi.BlockInsertionResult) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyUTXOsChanged")
	defer onEnd()

	utxoIndexChanges, err := m.context.UTXOIndex.Update(blockInsertionResult.SelectedParentChainChanges)
	if err != nil {
		return err
	}
	return m.context.NotificationManager.NotifyUTXOsChanged(utxoIndexChanges)
}

func (m *Manager) notifyVirtualSelectedParentBlueScoreChanged() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualSelectedParentBlueScoreChanged")
	defer onEnd()

	virtualInfo, err := m.context.Domain.Consensus().GetVirtualInfo()
	if err != nil {
		return err
	}
	notification := appmessage.NewVirtualSelectedParentBlueScoreChangedNotificationMessage(virtualInfo.BlueScore)
	return m.context.NotificationManager.NotifyVirtualSelectedParentBlueScoreChanged(notification)
}
