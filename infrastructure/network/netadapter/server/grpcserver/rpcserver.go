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

const rpcMaxInboundConnections = 8

// NewRPCServer creates a new RPCServer
func NewRPCServer(listeningAddresses []string) (server.Server, error) {
	gRPCServer := newGRPCServer(listeningAddresses, RPCMaxMessageSize, rpcMaxInboundConnections, "RPC")
	rpcServer := &rpcServer{gRPCServer: *gRPCServer}
	protowire.RegisterRPCServer(gRPCServer.server, rpcServer)
	return rpcServer, nil
}

func (r *rpcServer) MessageStream(stream protowire.RPC_MessageStreamServer) error {
	defer panics.HandlePanic(log, "rpcServer.MessageStream", nil)

	return r.handleInboundConnection(stream.Context(), stream)
}
