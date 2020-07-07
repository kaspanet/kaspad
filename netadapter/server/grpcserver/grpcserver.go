package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/wire"
)

type gRPCServer struct {
	peerConnectedHandler server.PeerConnectedHandler
	connections          []server.Connection
}

// NewGRPCServer creates and starts a gRPC server with the given
// listening port
func NewGRPCServer(listeningPort string) (server.Server, error) {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil, nil
}

// SetPeerConnectedHandler sets the peer connected handler
// function for the server
func (s *gRPCServer) SetPeerConnectedHandler(peerConnectedHandler server.PeerConnectedHandler) {
	s.peerConnectedHandler = peerConnectedHandler
}

// Connect connects to the given address
// This is part of the Server interface
func (s *gRPCServer) Connect(address string) (server.Connection, error) {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil, nil
}

// Connections returns a slice of connections the server
// is currently connected to.
// This is part of the Server interface
func (s *gRPCServer) Connections() []server.Connection {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil
}

func (s *gRPCServer) Close() error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil
}

type gRPCConnection struct{}

// Send sends the given message through the connection
// This is part of the Connection interface
func (c *gRPCConnection) Send(message wire.Message) error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil
}

// Receive receives the next message from the connection
// This is part of the Connection interface
func (c *gRPCConnection) Receive() (wire.Message, error) {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil, nil
}

// Disconnect disconnects the connection
// This is part of the Connection interface
func (c *gRPCConnection) Disconnect() error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
	return nil
}
