package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver"
	"sync/atomic"
)

// RouterInitializer is a function that initializes a new
// router to be used with a new connection
type RouterInitializer func() (*Router, error)

// NetAdapter is an abstraction layer over networking.
// This type expects a RouteInitializer function. This
// function weaves together the various "inputRoutes" (messages
// and message handlers) without exposing anything related
// to networking internals.
type NetAdapter struct {
	server            server.Server
	routerInitializer RouterInitializer
	stop              int32
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningAddrs []string) (*NetAdapter, error) {
	s, err := grpcserver.NewGRPCServer(listeningAddrs)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		server: s,
	}

	onConnectedHandler := adapter.newOnConnectedHandler()
	adapter.server.SetOnConnectedHandler(onConnectedHandler)

	return &adapter, nil
}

// Start begins the operation of the NetAdapter
func (na *NetAdapter) Start() error {
	return na.server.Start()
}

// Stop safely closes the NetAdapter
func (na *NetAdapter) Stop() error {
	if atomic.AddInt32(&na.stop, 1) != 1 {
		log.Warnf("Net adapter stopped more than once")
		return nil
	}
	return na.server.Stop()
}

func (na *NetAdapter) newOnConnectedHandler() server.OnConnectedHandler {
	return func(serverConnection server.Connection) error {
		router, err := na.routerInitializer()
		if err != nil {
			return err
		}
		serverConnection.SetOnDisconnectedHandler(func() error {
			return router.Close()
		})

		na.startReceiveLoop(serverConnection, router)
		return nil
	}
}

func (na *NetAdapter) startReceiveLoop(serverConnection server.Connection, router *Router) {
	spawn(func() {
		for {
			if atomic.LoadInt32(&na.stop) != 0 {
				err := serverConnection.Disconnect()
				if err != nil {
					log.Warnf("Failed to disconnect from %s: %s", serverConnection, err)
				}
				return
			}

			message, err := serverConnection.Receive()
			if err != nil {
				log.Warnf("Received error from %s: %s", serverConnection, err)
				err := serverConnection.Disconnect()
				if err != nil {
					log.Warnf("Failed to disconnect from %s: %s", serverConnection, err)
				}
			}
			router.RouteMessage(message)
		}
	})
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer RouterInitializer) {
	na.routerInitializer = routerInitializer
}
