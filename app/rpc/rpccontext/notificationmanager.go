package rpccontext

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"sync"
)

type NotificationManager struct {
	sync.RWMutex
	listeners map[*router.Router]*NotificationListener
}

type OnBlockAddedListener func(block *util.Block) error
type OnChainChangedListener func(notification *appmessage.ChainChangedNotificationMessage) error

type NotificationListener struct {
	onBlockAddedListener           OnBlockAddedListener
	onBlockAddedNotificationChan   chan *util.Block
	onChainChangedListener         OnChainChangedListener
	onChainChangedNotificationChan chan *appmessage.ChainChangedNotificationMessage

	closeChan chan struct{}
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		listeners: make(map[*router.Router]*NotificationListener),
	}
}

func (nm *NotificationManager) AddListener(router *router.Router) *NotificationListener {
	nm.Lock()
	defer nm.Unlock()

	listener := NewNotificationListener()
	nm.listeners[router] = listener
	return listener
}

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

func (nm *NotificationManager) Listener(router *router.Router) (*NotificationListener, error) {
	nm.RLock()
	defer nm.RUnlock()

	listener, ok := nm.listeners[router]
	if !ok {
		return nil, errors.Errorf("listener not found")
	}
	return listener, nil
}

func (nm *NotificationManager) NotifyBlockAdded(block *util.Block) {
	nm.RLock()
	defer nm.RUnlock()

	for _, listener := range nm.listeners {
		if listener.onBlockAddedListener != nil {
			select {
			case listener.onBlockAddedNotificationChan <- block:
			case <-listener.closeChan:
				continue
			}
		}
	}
}

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

func NewNotificationListener() *NotificationListener {
	return &NotificationListener{
		onBlockAddedNotificationChan: make(chan *util.Block),
		closeChan:                    make(chan struct{}, 1),
	}
}

func (nl *NotificationListener) SetOnBlockAddedListener(onBlockAddedListener OnBlockAddedListener) {
	nl.onBlockAddedListener = onBlockAddedListener
}

func (nl *NotificationListener) SetOnChainChangedListener(onChainChangedListener OnChainChangedListener) {
	nl.onChainChangedListener = onChainChangedListener
}

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
