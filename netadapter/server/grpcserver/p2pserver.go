package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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
	connection := newConnection(peerInfo.Addr)

	spawn(func() { testConnection(connection, peerInfo) })

	p.server.addConnection(connection)
	log.Infof("Incoming connection from %s", peerInfo.Addr)
	err := connection.serverConnectionLoop(stream)
	if err != nil {
		log.Errorf("Error in serverConnectionLoop: %+v", err)
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func testConnection(connection *gRPCConnection, peerInfo *peer.Peer) {
	msg, err := connection.Receive()
	if err != nil {
		log.Errorf("Error receiving from %s: %+v", peerInfo.Addr, err)
	}
	log.Infof("Got message from %s: %s", peerInfo.Addr, msg.Command())

	err = connection.Send(wire.NewMsgPong(667))
	if err != nil {
		log.Errorf("Error sending to %s: %+v", peerInfo.Addr, err)
	}
}
