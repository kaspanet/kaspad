package grpcserver

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
	"net"
)

type rpcServer struct {
	protowire.UnimplementedRPCServer
	gRPCServer
}

func NewRPCServer(listeningAddresses []string) (server.Server, error) {
	gRPCServer := newGRPCServer(listeningAddresses)
	rpcServer := &rpcServer{gRPCServer: *gRPCServer}
	protowire.RegisterRPCServer(gRPCServer.server, rpcServer)
	return rpcServer, nil
}

func (r *rpcServer) MessageStream(stream protowire.RPC_MessageStreamServer) error {
	defer panics.HandlePanic(log, "rpcServer.MessageStream", nil)

	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.Errorf("Error getting stream peer info from context")
	}
	tcpAddress, ok := peerInfo.Addr.(*net.TCPAddr)
	if !ok {
		return errors.Errorf("non-tcp connections are not supported")
	}

	connection := newConnection(&r.gRPCServer, tcpAddress, false, stream)

	err := r.onConnectedHandler(connection)
	if err != nil {
		return err
	}

	log.Infof("Incoming connection from %s", peerInfo.Addr)

	<-connection.stopChan

	return nil
}
