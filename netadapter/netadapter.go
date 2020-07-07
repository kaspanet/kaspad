package netadapter

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpc"
)

type netAdapter struct {
	routerInitializer func() router
	server            server.Server
}

func newNetAdapter(listeningPort string, routerInitializer func() router) (*netAdapter, error) {
	server, err := grpc.NewGRPCServer(listeningPort)
	if err != nil {
		return nil, err
	}

	adapter := netAdapter{
		routerInitializer: routerInitializer,
		server:            server,
	}
	return &adapter, nil
}
