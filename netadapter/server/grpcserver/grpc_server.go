package grpcserver

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/peer"

	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	onConnectedHandler server.OnConnectedHandler
	connections        []*gRPCConnection
	listeningAddrs     []string
	server             *grpc.Server
}

// NewGRPCServer creates and starts a gRPC server with the given
// listening port
func NewGRPCServer(listeningAddrs []string) (server.Server, error) {
	s := &gRPCServer{
		server:         grpc.NewServer(),
		listeningAddrs: listeningAddrs,
	}
	protowire.RegisterP2PServer(s.server, &p2pServer{})

	return s, nil
}

func (s *gRPCServer) Start() error {
	for _, listenAddr := range s.listeningAddrs {
		listener, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return errors.Wrapf(err, "Error listening on %s", listenAddr)
		}

		spawn(func() {
			err := s.server.Serve(listener)
			if err != nil {
				panics.Exit(log, fmt.Sprintf("Error serving on %s: %+v", listenAddr, err))
			}
		})

		log.Infof("P2P server listening on %s", listenAddr)
	}

	return nil
}

func (s *gRPCServer) Stop() error {
	s.server.GracefulStop()
	return nil
}

// SetOnConnectedHandler sets the peer connected handler
// function for the server
func (s *gRPCServer) SetOnConnectedHandler(onConnectedHandler server.OnConnectedHandler) {
	s.onConnectedHandler = onConnectedHandler
}

// Connect connects to the given address
// This is part of the Server interface
func (s *gRPCServer) Connect(address string) (server.Connection, error) {
	log.Infof("Dialing to %s", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "Error connecting to %s", address)
	}
	client := protowire.NewP2PClient(conn)
	stream, err := client.MessageStream(context.Background())

	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return nil, errors.Errorf("Error getting stream peer info from context")
	}

	connection := newConnection(peerInfo.Addr)
	spawn(func() { connection.clientConnectionLoop(stream) })
	return connection, nil
}

// Connections returns a slice of connections the server
// is currently connected to.
// This is part of the Server interface
func (s *gRPCServer) Connections() []server.Connection {
	result := make([]server.Connection, 0, len(s.connections))

	for _, conn := range s.connections {
		result = append(result, conn)
	}

	return result
}
