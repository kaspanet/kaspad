package grpcserver

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
)

type rpcServer struct {
	protowire.UnimplementedRPCServer
	gRPCServer
}

// RPCMaxMessageSize is the max message size for the RPC server to send and receive
const RPCMaxMessageSize = 1024 * 1024 * 1024 // 1 GB

// RPCMaxInboundConnections is the max amount of inbound connections for the RPC server
const RPCMaxInboundConnections = 128

// NewRPCServer creates a new RPCServer
func NewRPCServer(listeningAddresses []string) (server.Server, error) {
	gRPCServer := newGRPCServer(listeningAddresses, RPCMaxMessageSize, RPCMaxInboundConnections, "RPC")
	rpcServer := &rpcServer{gRPCServer: *gRPCServer}
	protowire.RegisterRPCServer(gRPCServer.server, rpcServer)
	return rpcServer, nil
}

func (r *rpcServer) MessageStream(stream protowire.RPC_MessageStreamServer) error {
	defer panics.HandlePanic(log, "rpcServer.MessageStream", nil)

	return r.handleInboundConnection(stream.Context(), stream)
}
