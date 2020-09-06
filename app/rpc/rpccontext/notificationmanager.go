package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
	"sync"
)

// NotificationManager manages notifications for the RPC
type NotificationManager struct {
	sync.RWMutex
	listeners map[*router.Router]*NotificationListener
}

// OnBlockAddedListener is a listener function for when a block is added to the DAG
type OnBlockAddedListener func(notification *appmessage.BlockAddedNotificationMessage) error

// OnChainChangedListener is a listener function for when the DAG's selected parent chain changes
type OnChainChangedListener func(notification *appmessage.ChainChangedNotificationMessage) error

// NotificationListener represents a registered RPC notification listener
type NotificationListener struct {
	onBlockAddedListener           OnBlockAddedListener
	onBlockAddedNotificationChan   chan *appmessage.BlockAddedNotificationMessage
	onChainChangedListener         OnChainChangedListener
	onChainChangedNotificationChan chan *appmessage.ChainChangedNotificationMessage

	closeChan chan struct{}
}

// NewNotificationManager creates a new NotificationManager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		listeners: make(map[*router.Router]*NotificationListener),
	}
}

// AddListener registers a listener with the given router
func (nm *NotificationManager) AddListener(router *router.Router) *NotificationListener {
	nm.Lock()
	defer nm.Unlock()

	listener := newNotificationListener()
	nm.listeners[router] = listener
	return listener
}

// RemoveListener unregisters the given router
func (nm *NotificationManager) RemoveListener(router *router.Router) {
	nm.Lock()
	defer nm.Unlock()

	listener, ok := nm.listeners[router]
	if !ok {
		return
	}
	listener.close()

	delete(nm.listeners, router)
}

// Listener retrieves the listener registered with the given router
func (nm *NotificationManager) Listener(router *router.Router) (*NotificationListener, error) {
	nm.RLock()
	defer nm.RUnlock()

	listener, ok := nm.listeners[router]
	if !ok {
		return nil, errors.Errorf("listener not found")
	}
	return listener, nil
}

// NotifyBlockAdded notifies the notification manager that a block has been added to the DAG
func (nm *NotificationManager) NotifyBlockAdded(notification *appmessage.BlockAddedNotificationMessage) {
	nm.RLock()
	defer nm.RUnlock()

	for _, listener := range nm.listeners {
		if listener.onBlockAddedListener != nil {
			select {
			case listener.onBlockAddedNotificationChan <- notification:
			case <-listener.closeChan:
				continue
			}
		}
	}
}

// NotifyChainChanged notifies the notification manager that the DAG's selected parent chain has changed
func (nm *NotificationManager) NotifyChainChanged(message *appmessage.ChainChangedNotificationMessage) {
	nm.RLock()
	defer nm.RUnlock()

	for _, listener := range nm.listeners {
		if listener.onChainChangedListener != nil {
			select {
			case listener.onChainChangedNotificationChan <- message:
			case <-listener.closeChan:
				continue
			}
		}
	}
}

func newNotificationListener() *NotificationListener {
	return &NotificationListener{
		onBlockAddedNotificationChan:   make(chan *appmessage.BlockAddedNotificationMessage),
		onChainChangedNotificationChan: make(chan *appmessage.ChainChangedNotificationMessage),
		closeChan:                      make(chan struct{}, 1),
	}
}

// SetOnBlockAddedListener sets the onBlockAddedListener handler for this listener
func (nl *NotificationListener) SetOnBlockAddedListener(onBlockAddedListener OnBlockAddedListener) {
	nl.onBlockAddedListener = onBlockAddedListener
}

// SetOnChainChangedListener sets the onChainChangedListener handler for this listener
func (nl *NotificationListener) SetOnChainChangedListener(onChainChangedListener OnChainChangedListener) {
	nl.onChainChangedListener = onChainChangedListener
}

// ProcessNextNotification waits until a notification arrives and processes it
func (nl *NotificationListener) ProcessNextNotification() error {
	select {
	case block := <-nl.onBlockAddedNotificationChan:
		return nl.onBlockAddedListener(block)
	case notification := <-nl.onChainChangedNotificationChan:
		return nl.onChainChangedListener(notification)
	case <-nl.closeChan:
		return nil
	}
}

func (nl *NotificationListener) close() {
	nl.closeChan <- struct{}{}
}
