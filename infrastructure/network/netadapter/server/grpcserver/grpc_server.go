package grpcserver

import (
	"fmt"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
)

type gRPCServer struct {
	onConnectedHandler server.OnConnectedHandler
	listeningAddresses []string
	server             *grpc.Server
}

const maxMessageSize = 1024 * 1024 * 10 // 10MB

// newGRPCServer creates a gRPC server
func newGRPCServer(listeningAddresses []string) *gRPCServer {
	return &gRPCServer{
		server:             grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize)),
		listeningAddresses: listeningAddresses,
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

	log.Debugf("Server started with maxMessageSize %d", maxMessageSize)

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

	log.Infof("Server listening on %s", listenAddr)
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
