package grpcserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/server"
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// RequestModifier can modify the http request
type RequestModifier func(r *grpc.Server)

type gRPCServer struct {
	// modifiers are applied before any request
	//modifiers          []RequestModifier
	onConnectedHandler server.OnConnectedHandler
	listeningAddresses []string
	server             *grpc.Server
	name               string
	auth               string

	maxInboundConnections      int
	inboundConnectionCount     int
	inboundConnectionCountLock *sync.Mutex
}

// newGRPCServer creates a gRPC server
func newGRPCServer(listeningAddresses []string, maxMessageSize int, maxInboundConnections int, name string, auth string, certFile string, keyFile string) *gRPCServer {
	log.Debugf("Created new %s GRPC server with maxMessageSize %d and maxInboundConnections %d", name, maxMessageSize, maxInboundConnections)
	log.Warnf("Name: %s for grpc auth type: %s", name, auth)
	if auth == "tls" {
		creds, _ := credentials.NewServerTLSFromFile(certFile, keyFile)
		return &gRPCServer{
			server:                     grpc.NewServer(grpc.Creds(creds), grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize)),
			listeningAddresses:         listeningAddresses,
			name:                       name,
			auth:                       auth,
			maxInboundConnections:      maxInboundConnections,
			inboundConnectionCount:     0,
			inboundConnectionCountLock: &sync.Mutex{},
		}
	} else {
		return &gRPCServer{
			server:                     grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize)),
			listeningAddresses:         listeningAddresses,
			name:                       name,
			auth:                       auth,
			maxInboundConnections:      maxInboundConnections,
			inboundConnectionCount:     0,
			inboundConnectionCountLock: &sync.Mutex{},
		}
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

	spawn(fmt.Sprintf("%s.gRPCServer.listenOn-Serve", s.name), func() {
		err := s.server.Serve(listener)
		if err != nil {
			panics.Exit(log, fmt.Sprintf("error serving %s on %s: %+v", s.name, listenAddr, err))
		}
	})

	log.Infof("%s Server listening on %s", s.name, listener.Addr())
	return nil
}

func (s *gRPCServer) Stop() error {
	const stopTimeout = 2 * time.Second

	stopChan := make(chan interface{})
	spawn("gRPCServer.Stop", func() {
		s.server.GracefulStop()
		close(stopChan)
	})

	select {
	case <-stopChan:
	case <-time.After(stopTimeout):
		log.Warnf("Could not gracefully stop %s: timed out after %s", s.name, stopTimeout)
		s.server.Stop()
	}
	return nil
}

// SetOnConnectedHandler sets the peer connected handler
// function for the server
func (s *gRPCServer) SetOnConnectedHandler(onConnectedHandler server.OnConnectedHandler) {
	s.onConnectedHandler = onConnectedHandler
}

func (s *gRPCServer) handleInboundConnection(ctx context.Context, stream grpcStream) error {
	connectionCount, err := s.incrementInboundConnectionCountAndLimitIfRequired()
	if err != nil {
		return err
	}
	defer s.decrementInboundConnectionCount()

	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return errors.Errorf("Error getting stream peer info from context")
	}
	tcpAddress, ok := peerInfo.Addr.(*net.TCPAddr)
	if !ok {
		return errors.Errorf("non-tcp connections are not supported")
	}

	connection := newConnection(s, tcpAddress, stream, nil)

	err = s.onConnectedHandler(connection)
	if err != nil {
		return err
	}

	log.Infof("%s Incoming connection from %s #%d", s.name, peerInfo.Addr, connectionCount)

	<-connection.stopChan
	return nil
}

func (s *gRPCServer) incrementInboundConnectionCountAndLimitIfRequired() (int, error) {
	s.inboundConnectionCountLock.Lock()
	defer s.inboundConnectionCountLock.Unlock()

	if s.maxInboundConnections > 0 && s.inboundConnectionCount == s.maxInboundConnections {
		log.Warnf("Limit of %d %s inbound connections has been exceeded", s.maxInboundConnections, s.name)
		return s.inboundConnectionCount, errors.Errorf("limit of %d %s inbound connections has been exceeded", s.maxInboundConnections, s.name)
	}

	s.inboundConnectionCount++
	return s.inboundConnectionCount, nil
}

func (s *gRPCServer) decrementInboundConnectionCount() {
	s.inboundConnectionCountLock.Lock()
	defer s.inboundConnectionCountLock.Unlock()

	s.inboundConnectionCount--
}
