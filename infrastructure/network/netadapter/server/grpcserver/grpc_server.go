package grpcserver

import (
	"context"
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
)

type gRPCServer struct {
	onConnectedHandler server.OnConnectedHandler
	listeningAddresses []string
	server             *grpc.Server
	name               string
}

// newGRPCServer creates a gRPC server
func newGRPCServer(listeningAddresses []string, maxMessageSize int, name string) *gRPCServer {
	log.Debugf("Created new GRPC server with maxMessageSize %d", maxMessageSize)
	return &gRPCServer{
		server:             grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize)),
		listeningAddresses: listeningAddresses,
		name:               name,
	}
}

func (s *gRPCServer) Start() error {
	if s.onConnectedHandler == nil {
		return errors.New("onConnectedHandler is nil")
	}

	for _, listenAddress := range s.listeningAddresses {
		err := s.listenOn(listenAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gRPCServer) listenOn(listenAddr string) error {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return errors.Wrapf(err, "%s error listening on %s", s.name, listenAddr)
	}

	spawn(s.name+" gRPCServer.listenOn-Serve", func() {
		err := s.server.Serve(listener)
		if err != nil {
			panics.Exit(log, fmt.Sprintf("error serving on %s: %+v", listenAddr, err))
		}
	})

	log.Infof("%s Server listening on %s", s.name, listenAddr)
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

func (s *gRPCServer) handleInboundConnection(ctx context.Context, stream grpcStream) error {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return errors.Errorf("Error getting stream peer info from context")
	}
	tcpAddress, ok := peerInfo.Addr.(*net.TCPAddr)
	if !ok {
		return errors.Errorf("non-tcp connections are not supported")
	}

	connection := newConnection(s, tcpAddress, stream, nil)

	err := s.onConnectedHandler(connection)
	if err != nil {
		return err
	}

	log.Infof("%s Incoming connection from %s", s.name, peerInfo.Addr)

	<-connection.stopChan

	return nil
}
