package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc/peer"

	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	onConnectedHandler server.OnConnectedHandler
	listeningAddrs     []string
	server             *grpc.Server
}

// NewGRPCServer creates and starts a gRPC server, listening on the
// provided addresses/ports
func NewGRPCServer(listeningAddrs []string) (server.Server, error) {
	s := &gRPCServer{
		server:         grpc.NewServer(),
		listeningAddrs: listeningAddrs,
	}
	protowire.RegisterP2PServer(s.server, newP2PServer(s))

	return s, nil
}

func (s *gRPCServer) Start() error {
	for _, listenAddr := range s.listeningAddrs {
		err := s.listenOn(listenAddr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gRPCServer) listenOn(listenAddr string) error {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return errors.Wrapf(err, "error listening on %s", listenAddr)
	}

	spawn("gRPCServer.listenOn-Serve", func() {
		err := s.server.Serve(listener)
		if err != nil {
			panics.Exit(log, fmt.Sprintf("error serving on %s: %+v", listenAddr, err))
		}
	})

	log.Infof("P2P server listening on %s", listenAddr)
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

	const dialTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	gRPCConnection, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to %s", address)
	}

	client := protowire.NewP2PClient(gRPCConnection)
	stream, err := client.MessageStream(context.Background())
	if err != nil {
		return nil, errors.Wrapf(err, "error getting client stream for %s", address)
	}

	peerInfo, ok := peer.FromContext(stream.Context())
	if !ok {
		return nil, errors.Errorf("error getting stream peer info from context for %s", address)
	}
	tcpAddress, ok := peerInfo.Addr.(*net.TCPAddr)
	if !ok {
		return nil, errors.Errorf("non-tcp addresses are not supported")
	}

	connection := newConnection(s, tcpAddress, true, stream)

	err = s.onConnectedHandler(connection)
	if err != nil {
		return nil, err
	}

	log.Infof("Connected to %s", address)

	return connection, nil
}
