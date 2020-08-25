package grpcserver

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protorpc"
)

type rpcServer struct {
	protorpc.UnimplementedRPCServer
	gRPCServer
}

func NewRPCServer(listeningAddresses []string) (server.Server, error) {
	gRPCServer := newGRPCServer(listeningAddresses)
	rpcServer := &rpcServer{gRPCServer: *gRPCServer}
	protorpc.RegisterRPCServer(gRPCServer.server, rpcServer)
	return rpcServer, nil
}
