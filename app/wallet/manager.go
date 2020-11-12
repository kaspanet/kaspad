package wallet

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc"
	"github.com/kaspanet/kaspad/app/wallet/wallethandlers"
	"github.com/kaspanet/kaspad/app/wallet/walletnotification"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

type sender interface {
	RegisterHandler(command appmessage.MessageCommand, rpcHandler rpc.Handler)
}

var walletHandlers = map[appmessage.MessageCommand]wallethandlers.HandlerFunc{
	appmessage.CmdNotifyBlockAddedRequestMessage:           wallethandlers.HandleNotifyBlockAdded,
	appmessage.CmdNotifyTransactionAddedRequestMessage:     wallethandlers.HandleNotifyTransactionAdded,
	appmessage.CmdNotifyUTXOOfAddressChangedRequestMessage: wallethandlers.HandleNotifyUTXOOfAddressChanged,
	appmessage.CmdNotifyFinalityConflictsRequestMessage:    wallethandlers.HandleNotifyFinalityConflicts,
}

// Manager is an wallet manager
type Manager struct {
	rpcManager          *rpc.Manager
	notificationManager *walletnotification.Manager
}

// NewManager creates a new wallet Manager
func NewManager(rpcManager *rpc.Manager) *Manager {
	return &Manager{
		rpcManager:          rpcManager,
		notificationManager: walletnotification.NewNotificationManager(),
	}
}

// RegisterWalletHandlers register all wallet manger handlers to the sender
func (m *Manager) RegisterWalletHandlers(handlerSender sender) {
	for command, handler := range walletHandlers {
		handlerInContext := wallethandlers.HandlerInContext{
			Context: m,
			Handler: handler,
		}
		handlerSender.RegisterHandler(command, &handlerInContext)
	}
}

// Listener retrieves the listener registered with the given router
func (m *Manager) Listener(router *routerpkg.Router) (*walletnotification.Listener, error) {
	return m.notificationManager.Listener(router)
}

// NotifyBlockAddedToDAG notifies the manager that a block has been added to the DAG
func (m *Manager) NotifyBlockAddedToDAG(block *externalapi.DomainBlock) error {
	notification := appmessage.NewBlockAddedNotificationMessage(appmessage.DomainBlockToMsgBlock(block))

	err := m.notificationManager.NotifyBlockAdded(notification)
	if err != nil {
		return err
	}

	err = m.notificationManager.NotifyTransactionAdded(block.Transactions)
	if err != nil {
		return err
	}

	return nil
}

// NotifyUTXOOfAddressChanged notifies the manager that a associated utxo set with address was changed
func (m *Manager) NotifyUTXOOfAddressChanged(addresses []string) error {
	notification := appmessage.NewUTXOOfAddressChangedNotificationMessage(addresses)
	return m.notificationManager.NotifyUTXOOfAddressChanged(notification)
}

// NotifyFinalityConflict notifies the manager that there's a finality conflict in the DAG
func (m *Manager) NotifyFinalityConflict(violatingBlockHash string) error {
	notification := appmessage.NewFinalityConflictNotificationMessage(violatingBlockHash)
	return m.notificationManager.NotifyFinalityConflict(notification)
}

// NotifyFinalityConflictResolved notifies the manager that a finality conflict in the DAG has been resolved
func (m *Manager) NotifyFinalityConflictResolved(finalityBlockHash string) error {
	notification := appmessage.NewFinalityConflictResolvedNotificationMessage(finalityBlockHash)
	return m.notificationManager.NotifyFinalityConflictResolved(notification)
}
