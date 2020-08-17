package grpcserver

import (
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
	"net"
)

type p2pServer struct {
	protowire.UnimplementedP2PServer
	server *gRPCServer
}

func newP2PServer(s *gRPCServer) *p2pServer {
	return &p2pServer{server: s}
}

func (p *p2pServer) MessageStream(stream protowire.P2P_MessageStreamServer) error {
	defer panics.HandlePanic(log, "p2pServer.MessageStream", nil)

	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.Errorf("Error getting stream peer info from context")
	}
	tcpAddress, ok := peerInfo.Addr.(*net.TCPAddr)
	if !ok {
		return errors.Errorf("non-tcp connections are not supported")
	}

	connection := newConnection(p.server, tcpAddress, false, stream)

	err := p.server.onConnectedHandler(connection)
	if err != nil {
		return err
	}

	log.Infof("Incoming connection from %s", peerInfo.Addr)

	<-connection.stopChan

	return nil
}
