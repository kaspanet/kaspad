package rpccontext

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/utxoindex"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
	"sync"
)

// NotificationManager manages notifications for the RPC
type NotificationManager struct {
	sync.RWMutex
	listeners map[*routerpkg.Router]*NotificationListener
}

// UTXOsChangedNotificationAddress represents a kaspad address.
// This type is meant to be used in UTXOsChanged notifications
type UTXOsChangedNotificationAddress struct {
	Address         string
	ScriptPublicKey []byte
}

// NotificationListener represents a registered RPC notification listener
type NotificationListener struct {
	propagateBlockAddedNotifications                            bool
	propagateVirtualSelectedParentChainChangedNotifications     bool
	propagateFinalityConflictNotifications                      bool
	propagateFinalityConflictResolvedNotifications              bool
	propagateUTXOsChangedNotifications                          bool
	propagateVirtualSelectedParentBlueScoreChangedNotifications bool

	propagateUTXOsChangedNotificationAddresses []*UTXOsChangedNotificationAddress
}

// NewNotificationManager creates a new NotificationManager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		listeners: make(map[*routerpkg.Router]*NotificationListener),
	}
}

// AddListener registers a listener with the given router
func (nm *NotificationManager) AddListener(router *routerpkg.Router) {
	nm.Lock()
	defer nm.Unlock()

	listener := newNotificationListener()
	nm.listeners[router] = listener
}

// RemoveListener unregisters the given router
func (nm *NotificationManager) RemoveListener(router *routerpkg.Router) {
	nm.Lock()
	defer nm.Unlock()

	delete(nm.listeners, router)
}

// Listener retrieves the listener registered with the given router
func (nm *NotificationManager) Listener(router *routerpkg.Router) (*NotificationListener, error) {
	nm.RLock()
	defer nm.RUnlock()

	listener, ok := nm.listeners[router]
	if !ok {
		return nil, errors.Errorf("listener not found")
	}
	return listener, nil
}

// NotifyBlockAdded notifies the notification manager that a block has been added to the DAG
func (nm *NotificationManager) NotifyBlockAdded(notification *appmessage.BlockAddedNotificationMessage) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateBlockAddedNotifications {
			err := router.OutgoingRoute().Enqueue(notification)
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				log.Warnf("Couldn't send notification: %s", err)
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyVirtualSelectedParentChainChanged notifies the notification manager that the DAG's selected parent chain has changed
func (nm *NotificationManager) NotifyVirtualSelectedParentChainChanged(notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateVirtualSelectedParentChainChangedNotifications {
			err := router.OutgoingRoute().Enqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyFinalityConflict notifies the notification manager that there's a finality conflict in the DAG
func (nm *NotificationManager) NotifyFinalityConflict(notification *appmessage.FinalityConflictNotificationMessage) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateFinalityConflictNotifications {
			err := router.OutgoingRoute().Enqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyFinalityConflictResolved notifies the notification manager that a finality conflict in the DAG has been resolved
func (nm *NotificationManager) NotifyFinalityConflictResolved(notification *appmessage.FinalityConflictResolvedNotificationMessage) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateFinalityConflictResolvedNotifications {
			err := router.OutgoingRoute().Enqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyUTXOsChanged notifies the notification manager that UTXOs have been changed
func (nm *NotificationManager) NotifyUTXOsChanged(utxoChanges *utxoindex.UTXOChanges) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateUTXOsChangedNotifications {
			// Filter utxoChanges and create a notification
			notification := listener.convertUTXOChangesToUTXOsChangedNotification(utxoChanges)

			// Don't send the notification if it's empty
			if len(notification.Added) == 0 && len(notification.Removed) == 0 {
				continue
			}

			// Enqueue the notification
			err := router.OutgoingRoute().Enqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyVirtualSelectedParentBlueScoreChanged notifies the notification manager that the DAG's
// virtual selected parent blue score has changed
func (nm *NotificationManager) NotifyVirtualSelectedParentBlueScoreChanged(
	notification *appmessage.VirtualSelectedParentBlueScoreChangedNotificationMessage) error {

	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateVirtualSelectedParentBlueScoreChangedNotifications {
			err := router.OutgoingRoute().Enqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newNotificationListener() *NotificationListener {
	return &NotificationListener{
		propagateBlockAddedNotifications:                            false,
		propagateVirtualSelectedParentChainChangedNotifications:     false,
		propagateFinalityConflictNotifications:                      false,
		propagateFinalityConflictResolvedNotifications:              false,
		propagateUTXOsChangedNotifications:                          false,
		propagateVirtualSelectedParentBlueScoreChangedNotifications: false,
	}
}

// PropagateBlockAddedNotifications instructs the listener to send block added notifications
// to the remote listener
func (nl *NotificationListener) PropagateBlockAddedNotifications() {
	nl.propagateBlockAddedNotifications = true
}

// PropagateVirtualSelectedParentChainChangedNotifications instructs the listener to send chain changed notifications
// to the remote listener
func (nl *NotificationListener) PropagateVirtualSelectedParentChainChangedNotifications() {
	nl.propagateVirtualSelectedParentChainChangedNotifications = true
}

// PropagateFinalityConflictNotifications instructs the listener to send finality conflict notifications
// to the remote listener
func (nl *NotificationListener) PropagateFinalityConflictNotifications() {
	nl.propagateFinalityConflictNotifications = true
}

// PropagateFinalityConflictResolvedNotifications instructs the listener to send finality conflict resolved notifications
// to the remote listener
func (nl *NotificationListener) PropagateFinalityConflictResolvedNotifications() {
	nl.propagateFinalityConflictResolvedNotifications = true
}

// PropagateUTXOsChangedNotifications instructs the listener to send UTXOs changed notifications
// to the remote listener
func (nl *NotificationListener) PropagateUTXOsChangedNotifications(addresses []*UTXOsChangedNotificationAddress) {
	nl.propagateUTXOsChangedNotifications = true
	nl.propagateUTXOsChangedNotificationAddresses = addresses
}

func (nl *NotificationListener) convertUTXOChangesToUTXOsChangedNotification(
	utxoChanges *utxoindex.UTXOChanges) *appmessage.UTXOsChangedNotificationMessage {

	notification := &appmessage.UTXOsChangedNotificationMessage{}
	for _, listenerAddress := range nl.propagateUTXOsChangedNotificationAddresses {
		listenerScriptPublicKeyHexString := utxoindex.ConvertScriptPublicKeyToHexString(listenerAddress.ScriptPublicKey)
		if addedPairs, ok := utxoChanges.Added[listenerScriptPublicKeyHexString]; ok {
			notification.Added = ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(listenerAddress.Address, addedPairs)
		}
		if removedOutpoints, ok := utxoChanges.Removed[listenerScriptPublicKeyHexString]; ok {
			for outpoint := range removedOutpoints {
				notification.Removed = append(notification.Removed, &appmessage.UTXOsByAddressesEntry{
					Address: listenerAddress.Address,
					Outpoint: &appmessage.RPCOutpoint{
						TransactionID: hex.EncodeToString(outpoint.TransactionID[:]),
						Index:         outpoint.Index,
					},
				})
			}
		}
	}
	return notification
}

// PropagateVirtualSelectedParentBlueScoreChangedNotifications instructs the listener to send
// virtual selected parent blue score notifications to the remote listener
func (nl *NotificationListener) PropagateVirtualSelectedParentBlueScoreChangedNotifications() {
	nl.propagateVirtualSelectedParentBlueScoreChangedNotifications = true
}
