package rpccontext

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
	"sync"
)

type NotificationManager struct {
	sync.RWMutex
	listeners map[*router.Router]*NotificationListener
}

type NotificationHandler func() error

type NotificationType int

const (
	BlockAdded NotificationType = iota
)

type NotificationListener struct {
	handlers    map[NotificationType]NotificationHandler
	handlerChan chan NotificationHandler
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		listeners: make(map[*router.Router]*NotificationListener),
	}
}

func (nm *NotificationManager) AddListener(router *router.Router) *NotificationListener {
	nm.Lock()
	defer nm.Unlock()

	listener := &NotificationListener{}
	nm.listeners[router] = listener
	return listener
}

func (nm *NotificationManager) RemoveListener(router *router.Router) {
	nm.Lock()
	defer nm.Unlock()

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

func (nm *NotificationManager) Notify(notificationType NotificationType) {
	nm.RLock()
	defer nm.RUnlock()

	for _, listener := range nm.listeners {
		handler, ok := listener.handlers[notificationType]
		if ok {
			listener.handlerChan <- handler
		}
	}
}

func (nl *NotificationListener) SetHandler(notificationType NotificationType, handler NotificationHandler) {
	nl.handlers[notificationType] = handler
}

func (nl *NotificationListener) NextHandler() NotificationHandler {
	return <-nl.handlerChan
}
