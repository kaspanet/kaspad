package rpccontext

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/dagconfig"

	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/utxoindex"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// DefaultNotificationID corrosponds to defualt grpc string value, and hence id value when not supplied, or as placeholder
const DefaultNotificationID string = ""

// NotificationManager manages notifications for the RPC
type NotificationManager struct {
	sync.RWMutex
	listeners map[*routerpkg.Router]*NotificationListener
	params    *dagconfig.Params
}

// UTXOsChangedNotificationAddress represents a kaspad address.
// This type is meant to be used in UTXOsChanged notifications
type UTXOsChangedNotificationAddress struct {
	Address               string
	ScriptPublicKeyString utxoindex.ScriptPublicKeyString
}

// NotificationListener represents a registered RPC notification listener
type NotificationListener struct {
	params *dagconfig.Params

	propagateBlockAddedNotifications                            bool
	propagateVirtualSelectedParentChainChangedNotifications     bool
	propagateFinalityConflictNotifications                      bool
	propagateFinalityConflictResolvedNotifications              bool
	propagateUTXOsChangedNotifications                          bool
	propagateVirtualSelectedParentBlueScoreChangedNotifications bool
	propagateVirtualDaaScoreChangedNotifications                bool
	propagatePruningPointUTXOSetOverrideNotifications           bool
	propagateNewBlockTemplateNotifications                      bool

	propagateBlockAddedNotificationsID                            string
	propagateVirtualSelectedParentChainChangedNotificationsID     string
	propagateFinalityConflictNotificationsID                      string
	propagateFinalityConflictResolvedNotificationsID              string
	propagateUTXOsChangedNotificationsID                          string
	propagateVirtualSelectedParentBlueScoreChangedNotificationsID string
	propagateVirtualDaaScoreChangedNotificationsID                string
	propagatePruningPointUTXOSetOverrideNotificationsID           string
	propagateNewBlockTemplateNotificationsID                      string

	propagateUTXOsChangedNotificationAddresses                                    map[utxoindex.ScriptPublicKeyString]*UTXOsChangedNotificationAddress
	includeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications bool
}

// NewNotificationManager creates a new NotificationManager
func NewNotificationManager(params *dagconfig.Params) *NotificationManager {
	return &NotificationManager{
		params:    params,
		listeners: make(map[*routerpkg.Router]*NotificationListener),
	}
}

// AddListener registers a listener with the given router
func (nm *NotificationManager) AddListener(router *routerpkg.Router) {
	nm.Lock()
	defer nm.Unlock()

	listener := newNotificationListener(nm.params)
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

// HasBlockAddedListeners indicates if the notification manager has any listeners for `BlockAdded` events
func (nm *NotificationManager) HasBlockAddedListeners() bool {
	nm.RLock()
	defer nm.RUnlock()

	for _, listener := range nm.listeners {
		if listener.propagateBlockAddedNotifications {
			return true
		}
	}
	return false
}

// NotifyBlockAdded notifies the notification manager that a block has been added to the DAG
func (nm *NotificationManager) NotifyBlockAdded(notification *appmessage.BlockAddedNotificationMessage) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateBlockAddedNotifications {

			notification.ID = listener.propagateBlockAddedNotificationsID

			err := router.OutgoingRoute().MaybeEnqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyVirtualSelectedParentChainChanged notifies the notification manager that the DAG's selected parent chain has changed
func (nm *NotificationManager) NotifyVirtualSelectedParentChainChanged(
	notification *appmessage.VirtualSelectedParentChainChangedNotificationMessage) error {

	nm.RLock()
	defer nm.RUnlock()

	notificationWithoutAcceptedTransactionIDs := &appmessage.VirtualSelectedParentChainChangedNotificationMessage{
		ID:                      DefaultNotificationID,
		RemovedChainBlockHashes: notification.RemovedChainBlockHashes,
		AddedChainBlockHashes:   notification.AddedChainBlockHashes,
	}

	for router, listener := range nm.listeners {
		if listener.propagateVirtualSelectedParentChainChangedNotifications {
			var err error

			if listener.includeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications {
				notification.ID = listener.propagateVirtualSelectedParentChainChangedNotificationsID
				err = router.OutgoingRoute().MaybeEnqueue(notification)
			} else {
				notificationWithoutAcceptedTransactionIDs.ID = listener.propagateVirtualSelectedParentChainChangedNotificationsID
				err = router.OutgoingRoute().MaybeEnqueue(notificationWithoutAcceptedTransactionIDs)
			}

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AllListenersThatPropagateVirtualSelectedParentChainChanged returns true if there's any listener that is
// subscribed to VirtualSelectedParentChainChanged notifications.
func (nm *NotificationManager) AllListenersThatPropagateVirtualSelectedParentChainChanged() []*NotificationListener {
	var listenersThatPropagate []*NotificationListener
	for _, listener := range nm.listeners {
		if listener.propagateVirtualSelectedParentChainChangedNotifications {
			listenersThatPropagate = append(listenersThatPropagate, listener)
		}
	}
	return listenersThatPropagate
}

// NotifyFinalityConflict notifies the notification manager that there's a finality conflict in the DAG
func (nm *NotificationManager) NotifyFinalityConflict(notification *appmessage.FinalityConflictNotificationMessage) error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateFinalityConflictNotifications {

			notification.ID = listener.propagateFinalityConflictNotificationsID

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

			notification.ID = listener.propagateFinalityConflictResolvedNotificationsID

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
			notification, err := listener.convertUTXOChangesToUTXOsChangedNotification(utxoChanges, listener.propagateUTXOsChangedNotificationsID)
			if err != nil {
				return err
			}

			// Don't send the notification if it's empty
			if len(notification.Added) == 0 && len(notification.Removed) == 0 {
				continue
			}

			// Enqueue the notification
			err = router.OutgoingRoute().MaybeEnqueue(notification)
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

			notification.ID = listener.propagateVirtualSelectedParentBlueScoreChangedNotificationsID

			err := router.OutgoingRoute().MaybeEnqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyVirtualDaaScoreChanged notifies the notification manager that the DAG's
// virtual DAA score has changed
func (nm *NotificationManager) NotifyVirtualDaaScoreChanged(
	notification *appmessage.VirtualDaaScoreChangedNotificationMessage) error {

	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateVirtualDaaScoreChangedNotifications {

			notification.ID = listener.propagateVirtualDaaScoreChangedNotificationsID

			err := router.OutgoingRoute().MaybeEnqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyNewBlockTemplate notifies the notification manager that a new
// block template is available for miners
func (nm *NotificationManager) NotifyNewBlockTemplate(
	notification *appmessage.NewBlockTemplateNotificationMessage) error {

	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagateNewBlockTemplateNotifications {

			notification.ID = listener.propagateNewBlockTemplateNotificationsID

			err := router.OutgoingRoute().Enqueue(notification)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotifyPruningPointUTXOSetOverride notifies the notification manager that the UTXO index
// reset due to pruning point change via IBD.
func (nm *NotificationManager) NotifyPruningPointUTXOSetOverride() error {
	nm.RLock()
	defer nm.RUnlock()

	for router, listener := range nm.listeners {
		if listener.propagatePruningPointUTXOSetOverrideNotifications {
			err := router.OutgoingRoute().Enqueue(appmessage.NewPruningPointUTXOSetOverrideNotificationMessage(listener.propagatePruningPointUTXOSetOverrideNotificationsID))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newNotificationListener(params *dagconfig.Params) *NotificationListener {
	return &NotificationListener{
		params: params,

		propagateBlockAddedNotifications:                            false,
		propagateVirtualSelectedParentChainChangedNotifications:     false,
		propagateFinalityConflictNotifications:                      false,
		propagateFinalityConflictResolvedNotifications:              false,
		propagateUTXOsChangedNotifications:                          false,
		propagateVirtualSelectedParentBlueScoreChangedNotifications: false,
		propagateVirtualDaaScoreChangedNotifications:                false,
		propagateNewBlockTemplateNotifications:                      false,
		propagatePruningPointUTXOSetOverrideNotifications:           false,
	}
}

// IncludeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications returns true if this listener
// includes accepted transaction IDs in it's virtual-selected-parent-chain-changed notifications
func (nl *NotificationListener) IncludeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications() bool {
	return nl.includeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications
}

// PropagateBlockAddedNotifications instructs the listener to send block added notifications
// to the remote listener
func (nl *NotificationListener) PropagateBlockAddedNotifications(id string) {
	nl.propagateBlockAddedNotificationsID = id
	nl.propagateBlockAddedNotifications = true
}

// PropagateVirtualSelectedParentChainChangedNotifications instructs the listener to send chain changed notifications
// to the remote listener
func (nl *NotificationListener) PropagateVirtualSelectedParentChainChangedNotifications(includeAcceptedTransactionIDs bool, id string) {
	nl.propagateVirtualSelectedParentChainChangedNotificationsID = id
	nl.propagateVirtualSelectedParentChainChangedNotifications = true
	nl.includeAcceptedTransactionIDsInVirtualSelectedParentChainChangedNotifications = includeAcceptedTransactionIDs
}

// PropagateFinalityConflictNotifications instructs the listener to send finality conflict notifications
// to the remote listener
func (nl *NotificationListener) PropagateFinalityConflictNotifications(id string) {
	nl.propagateFinalityConflictNotificationsID = id
	nl.propagateFinalityConflictNotifications = true
}

// PropagateFinalityConflictResolvedNotifications instructs the listener to send finality conflict resolved notifications
// to the remote listener
func (nl *NotificationListener) PropagateFinalityConflictResolvedNotifications(id string) {
	nl.propagateFinalityConflictResolvedNotificationsID = id
	nl.propagateFinalityConflictResolvedNotifications = true
}

// PropagateUTXOsChangedNotifications instructs the listener to send UTXOs changed notifications
// to the remote listener for the given addresses. Subsequent calls instruct the listener to
// send UTXOs changed notifications for those addresses along with the old ones. Duplicate addresses
// are ignored.
func (nl *NotificationListener) PropagateUTXOsChangedNotifications(addresses []*UTXOsChangedNotificationAddress, id string) {
	if !nl.propagateUTXOsChangedNotifications {
		nl.propagateUTXOsChangedNotificationsID = id
		nl.propagateUTXOsChangedNotifications = true
		nl.propagateUTXOsChangedNotificationAddresses =
			make(map[utxoindex.ScriptPublicKeyString]*UTXOsChangedNotificationAddress, len(addresses))
	}

	for _, address := range addresses {
		nl.propagateUTXOsChangedNotificationAddresses[address.ScriptPublicKeyString] = address
	}
}

// StopPropagatingUTXOsChangedNotifications instructs the listener to stop sending UTXOs
// changed notifications to the remote listener for the given addresses. Addresses for which
// notifications are not currently sent are ignored.
func (nl *NotificationListener) StopPropagatingUTXOsChangedNotifications(addresses []*UTXOsChangedNotificationAddress) {
	if !nl.propagateUTXOsChangedNotifications {
		return
	}

	for _, address := range addresses {
		delete(nl.propagateUTXOsChangedNotificationAddresses, address.ScriptPublicKeyString)
	}
}

func (nl *NotificationListener) convertUTXOChangesToUTXOsChangedNotification(
	utxoChanges *utxoindex.UTXOChanges, id string) (*appmessage.UTXOsChangedNotificationMessage, error) {

	// As an optimization, we iterate over the smaller set (O(n)) among the two below
	// and check existence over the larger set (O(1))
	utxoChangesSize := len(utxoChanges.Added) + len(utxoChanges.Removed)
	addressesSize := len(nl.propagateUTXOsChangedNotificationAddresses)

	notification := &appmessage.UTXOsChangedNotificationMessage{ID: id}

	if utxoChangesSize < addressesSize {
		for scriptPublicKeyString, addedPairs := range utxoChanges.Added {
			if listenerAddress, ok := nl.propagateUTXOsChangedNotificationAddresses[scriptPublicKeyString]; ok {
				utxosByAddressesEntries := ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(listenerAddress.Address, addedPairs)
				notification.Added = append(notification.Added, utxosByAddressesEntries...)
			}
		}
		for scriptPublicKeyString, removedPairs := range utxoChanges.Removed {
			if listenerAddress, ok := nl.propagateUTXOsChangedNotificationAddresses[scriptPublicKeyString]; ok {
				utxosByAddressesEntries := ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(listenerAddress.Address, removedPairs)
				notification.Removed = append(notification.Removed, utxosByAddressesEntries...)
			}
		}
	} else if addressesSize > 0 {
		for _, listenerAddress := range nl.propagateUTXOsChangedNotificationAddresses {
			listenerScriptPublicKeyString := listenerAddress.ScriptPublicKeyString
			if addedPairs, ok := utxoChanges.Added[listenerScriptPublicKeyString]; ok {
				utxosByAddressesEntries := ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(listenerAddress.Address, addedPairs)
				notification.Added = append(notification.Added, utxosByAddressesEntries...)
			}
			if removedPairs, ok := utxoChanges.Removed[listenerScriptPublicKeyString]; ok {
				utxosByAddressesEntries := ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(listenerAddress.Address, removedPairs)
				notification.Removed = append(notification.Removed, utxosByAddressesEntries...)
			}
		}
	} else {
		for scriptPublicKeyString, addedPairs := range utxoChanges.Added {
			addressString, err := nl.scriptPubKeyStringToAddressString(scriptPublicKeyString)
			if err != nil {
				return nil, err
			}

			utxosByAddressesEntries := ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(addressString, addedPairs)
			notification.Added = append(notification.Added, utxosByAddressesEntries...)
		}
		for scriptPublicKeyString, removedPAirs := range utxoChanges.Removed {
			addressString, err := nl.scriptPubKeyStringToAddressString(scriptPublicKeyString)
			if err != nil {
				return nil, err
			}

			utxosByAddressesEntries := ConvertUTXOOutpointEntryPairsToUTXOsByAddressesEntries(addressString, removedPAirs)
			notification.Removed = append(notification.Removed, utxosByAddressesEntries...)
		}
	}

	return notification, nil
}

func (nl *NotificationListener) scriptPubKeyStringToAddressString(scriptPublicKeyString utxoindex.ScriptPublicKeyString) (string, error) {
	scriptPubKey := utxoindex.ConvertStringToScriptPublicKey(scriptPublicKeyString)

	// ignore error because it is often returned when the script is of unknown type
	scriptType, address, err := txscript.ExtractScriptPubKeyAddress(scriptPubKey, nl.params)
	if err != nil {
		return "", err
	}

	var addressString string
	if scriptType == txscript.NonStandardTy {
		addressString = ""
	} else {
		addressString = address.String()
	}
	return addressString, nil
}

// PropagateVirtualSelectedParentBlueScoreChangedNotifications instructs the listener to send
// virtual selected parent blue score notifications to the remote listener
func (nl *NotificationListener) PropagateVirtualSelectedParentBlueScoreChangedNotifications(id string) {
	nl.propagateVirtualSelectedParentBlueScoreChangedNotificationsID = id
	nl.propagateVirtualSelectedParentBlueScoreChangedNotifications = true
}

// PropagateVirtualDaaScoreChangedNotifications instructs the listener to send
// virtual DAA score notifications to the remote listener
func (nl *NotificationListener) PropagateVirtualDaaScoreChangedNotifications(id string) {
	nl.propagateVirtualDaaScoreChangedNotificationsID = id
	nl.propagateVirtualDaaScoreChangedNotifications = true
}

// PropagateNewBlockTemplateNotifications instructs the listener to send
// new block template notifications to the remote listener
func (nl *NotificationListener) PropagateNewBlockTemplateNotifications(id string) {
	nl.propagateNewBlockTemplateNotificationsID = id
	nl.propagateNewBlockTemplateNotifications = true
}

// PropagatePruningPointUTXOSetOverrideNotifications instructs the listener to send pruning point UTXO set override notifications
// to the remote listener.
func (nl *NotificationListener) PropagatePruningPointUTXOSetOverrideNotifications(id string) {
	nl.propagatePruningPointUTXOSetOverrideNotificationsID = id
	nl.propagatePruningPointUTXOSetOverrideNotifications = true
}

// StopPropagatingPruningPointUTXOSetOverrideNotifications instructs the listener to stop sending pruning
// point UTXO set override notifications to the remote listener.
func (nl *NotificationListener) StopPropagatingPruningPointUTXOSetOverrideNotifications(id string) {
	nl.propagatePruningPointUTXOSetOverrideNotificationsID = id
	nl.propagatePruningPointUTXOSetOverrideNotifications = false
}
