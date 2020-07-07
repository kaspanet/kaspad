package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpc"
)

// NetAdapter is an adapter to the net
type NetAdapter struct {
	server            server.Server
	routerInitializer func() *Router
}

// NewNetAdapter creates and starts a new NetAdapter on the
// given listeningPort
func NewNetAdapter(listeningPort string) (*NetAdapter, error) {
	server, err := grpc.NewGRPCServer(listeningPort)
	if err != nil {
		return nil, err
	}
	adapter := NetAdapter{
		server: server,
	}
	return &adapter, nil
}

// SetRouterInitializer sets the routerInitializer function
// for the net adapter
func (na *NetAdapter) SetRouterInitializer(routerInitializer func() *Router) {
	na.routerInitializer = routerInitializer
}

func (na *NetAdapter) Close() error {
	return na.server.Close()
}
