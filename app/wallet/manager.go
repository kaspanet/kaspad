package wallet

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc"
	"github.com/kaspanet/kaspad/app/wallet/wallethandlers"
	"github.com/kaspanet/kaspad/app/wallet/walletnotification"
	"github.com/kaspanet/kaspad/domain/addressindex"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

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
func (m *Manager) RegisterWalletHandlers() {
	for command, handler := range walletHandlers {
		handlerInContext := wallethandlers.HandlerInContext{
			Context: m,
			Handler: handler,
		}
		m.rpcManager.SetNotifier(m.notificationManager)
		m.rpcManager.RegisterHandler(command, &handlerInContext)
	}
}

// Listener retrieves the listener registered with the given router
func (m *Manager) Listener(router *router.Router) (*walletnotification.Listener, error) {
	return m.notificationManager.Listener(router)
}

// NotifyBlockAddedToDAG notifies the manager that a block has been added to the DAG
func (m *Manager) NotifyBlockAddedToDAG(block *externalapi.DomainBlock) error {
	notification := appmessage.NewBlockAddedNotificationMessage(appmessage.DomainBlockToMsgBlock(block))

	err := m.notificationManager.NotifyBlockAdded(notification)
	if err != nil {
		return err
	}

	return nil
}

// NotifyTransactionAddedToDAG notifies the manager that a transaction has been added
func (m *Manager) NotifyTransactionAdded(transaction *externalapi.DomainTransaction, status externalapi.TransactionStatus, blueScore uint64, blockHash *externalapi.DomainHash, prefix util.Bech32Prefix) error {
	addresses, utxos, err := addressindex.GetAddressesAndUTXOsFromTransaction(transaction, blueScore, prefix)
	if err != nil {
		return err
	}
	utxosVerboseData := make([]*appmessage.UTXOVerboseData, 0, len(utxos))
	for _, utxoEntry := range utxos {
		utxoVerboseData := &appmessage.UTXOVerboseData{
			Amount:         utxoEntry.Amount,
			ScriptPubKey:   utxoEntry.ScriptPublicKey,
			BlockBlueScore: utxoEntry.BlockBlueScore,
			IsCoinbase:     utxoEntry.IsCoinbase,
		}
		utxosVerboseData = append(utxosVerboseData, utxoVerboseData)
	}

	notification := appmessage.NewTransactionAddedNotificationMessage(addresses, blockHash, utxosVerboseData, appmessage.DomainTransactionToMsgTx(transaction), status)
	err = m.notificationManager.NotifyTransactionAdded(notification)
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
