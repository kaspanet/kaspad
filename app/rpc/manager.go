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
func (m *Manager) NotifyBlockAddedToDAG(block *externalapi.DomainBlock, virtualChangeSet *externalapi.VirtualChangeSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyBlockAddedToDAG")
	defer onEnd()

	err := m.NotifyVirtualChange(virtualChangeSet)
	if err != nil {
		return err
	}

	rpcBlock := appmessage.DomainBlockToRPCBlock(block)
	err = m.context.PopulateBlockWithVerboseData(rpcBlock, block.Header, block, false)
	if err != nil {
		return err
	}
	blockAddedNotification := appmessage.NewBlockAddedNotificationMessage(rpcBlock)
	return m.context.NotificationManager.NotifyBlockAdded(blockAddedNotification)
}

// NotifyVirtualChange notifies the manager that the virtual block has been changed.
func (m *Manager) NotifyVirtualChange(virtualChangeSet *externalapi.VirtualChangeSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualChange")
	defer onEnd()

	if m.context.Config.UTXOIndex {
		err := m.notifyUTXOsChanged(virtualChangeSet)
		if err != nil {
			return err
		}
	}

	err := m.notifyVirtualSelectedParentBlueScoreChanged()
	if err != nil {
		return err
	}

	err = m.notifyVirtualDaaScoreChanged()
	if err != nil {
		return err
	}

	err = m.notifyVirtualSelectedParentChainChanged(virtualChangeSet)
	if err != nil {
		return err
	}

	return nil
}

// NotifyNewBlockTemplate notifies the manager that a new
// block template is available for miners
func (m *Manager) NotifyNewBlockTemplate() error {
	notification := appmessage.NewNewBlockTemplateNotificationMessage()
	return m.context.NotificationManager.NotifyNewBlockTemplate(notification)
}

// NotifyPruningPointUTXOSetOverride notifies the manager whenever the UTXO index
// resets due to pruning point change via IBD.
func (m *Manager) NotifyPruningPointUTXOSetOverride() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyPruningPointUTXOSetOverride")
	defer onEnd()

	if m.context.Config.UTXOIndex {
		err := m.notifyPruningPointUTXOSetOverride()
		if err != nil {
			return err
		}
	}

	return nil
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

func (m *Manager) notifyUTXOsChanged(virtualChangeSet *externalapi.VirtualChangeSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyUTXOsChanged")
	defer onEnd()

	utxoIndexChanges, err := m.context.UTXOIndex.Update(virtualChangeSet)
	if err != nil {
		return err
	}
	return m.context.NotificationManager.NotifyUTXOsChanged(utxoIndexChanges)
}

func (m *Manager) notifyPruningPointUTXOSetOverride() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.notifyPruningPointUTXOSetOverride")
	defer onEnd()

	err := m.context.UTXOIndex.Reset()
	if err != nil {
		return err
	}

	return m.context.NotificationManager.NotifyPruningPointUTXOSetOverride()
}

func (m *Manager) notifyVirtualSelectedParentBlueScoreChanged() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualSelectedParentBlueScoreChanged")
	defer onEnd()

	virtualSelectedParent, err := m.context.Domain.Consensus().GetVirtualSelectedParent()
	if err != nil {
		return err
	}

	blockInfo, err := m.context.Domain.Consensus().GetBlockInfo(virtualSelectedParent)
	if err != nil {
		return err
	}

	notification := appmessage.NewVirtualSelectedParentBlueScoreChangedNotificationMessage(blockInfo.BlueScore)
	return m.context.NotificationManager.NotifyVirtualSelectedParentBlueScoreChanged(notification)
}

func (m *Manager) notifyVirtualDaaScoreChanged() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualDaaScoreChanged")
	defer onEnd()

	virtualDAAScore, err := m.context.Domain.Consensus().GetVirtualDAAScore()
	if err != nil {
		return err
	}

	notification := appmessage.NewVirtualDaaScoreChangedNotificationMessage(virtualDAAScore)
	return m.context.NotificationManager.NotifyVirtualDaaScoreChanged(notification)
}

func (m *Manager) notifyVirtualSelectedParentChainChanged(virtualChangeSet *externalapi.VirtualChangeSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualSelectedParentChainChanged")
	defer onEnd()

	notification, err := m.context.ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(
		virtualChangeSet.VirtualSelectedParentChainChanges)
	if err != nil {
		return err
	}
	return m.context.NotificationManager.NotifyVirtualSelectedParentChainChanged(notification)
}
