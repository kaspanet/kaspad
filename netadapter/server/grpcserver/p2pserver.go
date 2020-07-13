package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
)

type p2pServer struct {
	protowire.UnimplementedP2PServer
	server *gRPCServer
}

func newP2PServer(s *gRPCServer) *p2pServer {
	return &p2pServer{server: s}
}

func (p *p2pServer) MessageStream(stream protowire.P2P_MessageStreamServer) error {
	defer panics.HandlePanic(log, nil)

	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return errors.Errorf("Error getting stream peer info from context")
	}
	connection := newConnection(p.server, peerInfo.Addr, false, stream)

	err := p.server.onConnectedHandler(connection)
	if err != nil {
		return err
	}

	log.Infof("Incoming connection from %s", peerInfo.Addr)

	<-connection.stopChan

	return nil
}
