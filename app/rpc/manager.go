package rpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/utxoindex"
	"github.com/kaspanet/kaspad/domain/txindex"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/addressmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/connmanager"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"
	"github.com/pkg/errors"
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
	txIndex *txindex.TXIndex,
	consensusEventsChan chan externalapi.ConsensusEvent,
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
			txIndex,
			shutDownChan,
		),
	}
	netAdapter.SetRPCRouterInitializer(manager.routerInitializer)

	manager.initConsensusEventsHandler(consensusEventsChan)

	return &manager
}

func (m *Manager) initConsensusEventsHandler(consensusEventsChan chan externalapi.ConsensusEvent) {
	spawn("consensusEventsHandler", func() {
		for {
			consensusEvent, ok := <-consensusEventsChan
			if !ok {
				return
			}
			switch event := consensusEvent.(type) {
			case *externalapi.VirtualChangeSet:
				err := m.notifyVirtualChange(event)
				if err != nil {
					panic(err)
				}
			case *externalapi.BlockAdded:
				err := m.notifyBlockAddedToDAG(event.Block)
				if err != nil {
					panic(err)
				}
			default:
				panic(errors.Errorf("Got event of unsupported type %T", consensusEvent))
			}
		}
	})
}

// notifyBlockAddedToDAG notifies the manager that a block has been added to the DAG
func (m *Manager) notifyBlockAddedToDAG(block *externalapi.DomainBlock) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.notifyBlockAddedToDAG")
	defer onEnd()

	// Before converting the block and populating it, we check if any listeners are interested.
	// This is done since most nodes do not use this event.
	if !m.context.NotificationManager.HasBlockAddedListeners() {
		return nil
	}

	rpcBlock := appmessage.DomainBlockToRPCBlock(block)
	err := m.context.PopulateBlockWithVerboseData(rpcBlock, block.Header, block, true)
	if err != nil {
		return err
	}
	blockAddedNotification := appmessage.NewBlockAddedNotificationMessage(rpcBlock)
	err = m.context.NotificationManager.NotifyBlockAdded(blockAddedNotification)
	if err != nil {
		return err
	}

	return nil
}

// notifyVirtualChange notifies the manager that the virtual block has been changed.
func (m *Manager) notifyVirtualChange(virtualChangeSet *externalapi.VirtualChangeSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualChange")
	defer onEnd()

	if m.context.Config.UTXOIndex && virtualChangeSet.VirtualUTXODiff != nil {
		err := m.notifyUTXOsChanged(virtualChangeSet)
		if err != nil {
			return err
		}
	}

	err := m.notifyVirtualSelectedParentBlueScoreChanged(virtualChangeSet.VirtualSelectedParentBlueScore)
	if err != nil {
		return err
	}

	err = m.notifyVirtualDaaScoreChanged(virtualChangeSet.VirtualDAAScore)
	if err != nil {
		return err
	}

	if virtualChangeSet.VirtualSelectedParentChainChanges == nil ||
		(len(virtualChangeSet.VirtualSelectedParentChainChanges.Added) == 0 &&
			len(virtualChangeSet.VirtualSelectedParentChainChanges.Removed) == 0) {

		return nil
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

func (m *Manager) notifyVirtualSelectedParentBlueScoreChanged(virtualSelectedParentBlueScore uint64) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualSelectedParentBlueScoreChanged")
	defer onEnd()

	notification := appmessage.NewVirtualSelectedParentBlueScoreChangedNotificationMessage(virtualSelectedParentBlueScore)
	return m.context.NotificationManager.NotifyVirtualSelectedParentBlueScoreChanged(notification)
}

func (m *Manager) notifyVirtualDaaScoreChanged(virtualDAAScore uint64) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualDaaScoreChanged")
	defer onEnd()

	notification := appmessage.NewVirtualDaaScoreChangedNotificationMessage(virtualDAAScore)
	return m.context.NotificationManager.NotifyVirtualDaaScoreChanged(notification)
}

func (m *Manager) notifyVirtualSelectedParentChainChanged(virtualChangeSet *externalapi.VirtualChangeSet) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "RPCManager.NotifyVirtualSelectedParentChainChanged")
	defer onEnd()

	listenersThatPropagateSelectedParentChanged :=
		m.context.NotificationManager.AllListenersThatPropagateVirtualSelectedParentChainChanged()
	if len(listenersThatPropagateSelectedParentChanged) > 0 {
		// Generating acceptedTransactionIDs is a heavy operation, so we check if it's needed by any listener.
		includeAcceptedTransactionIDs := false
		for _, listener := range listenersThatPropagateSelectedParentChanged {
			if listener.IncludeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications() {
				includeAcceptedTransactionIDs = true
				break
			}
		}

		notification, err := m.context.ConvertVirtualSelectedParentChainChangesToChainChangedNotificationMessage(
			virtualChangeSet.VirtualSelectedParentChainChanges, includeAcceptedTransactionIDs)
		if err != nil {
			return err
		}
		return m.context.NotificationManager.NotifyVirtualSelectedParentChainChanged(notification)
	}

	return nil
}
