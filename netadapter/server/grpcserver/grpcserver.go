package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/wire"
)

type gRPCServer struct {
	onConnectedHandler server.OnConnectedHandler
	connections        []server.Connection
}

// NewGRPCServer creates and starts a gRPC server with the given
// listening port
func NewGRPCServer(listeningAddrs []string) (server.Server, error) {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

func (s *gRPCServer) Start() error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

func (s *gRPCServer) Stop() error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

// SetOnConnectedHandler sets the on-connected handler
// function for the server
func (s *gRPCServer) SetOnConnectedHandler(onConnectedHandler server.OnConnectedHandler) {
	s.onConnectedHandler = onConnectedHandler
}

// Connect connects to the given address
// This is part of the Server interface
func (s *gRPCServer) Connect(address string) (server.Connection, error) {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

// Connections returns a slice of connections the server
// is currently connected to.
// This is part of the Server interface
func (s *gRPCServer) Connections() []server.Connection {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

type gRPCConnection struct {
	onDisconnectedHandler server.OnDisconnectedHandler
}

// Send sends the given message through the connection
// This is part of the Connection interface
func (c *gRPCConnection) Send(message wire.Message) error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

// Receive receives the next message from the connection
// This is part of the Connection interface
func (c *gRPCConnection) Receive() (wire.Message, error) {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

// Disconnect disconnects the connection
// This is part of the Connection interface
func (c *gRPCConnection) Disconnect() error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

func (c *gRPCConnection) IsConnected() bool {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}

// SetOnDisconnectedHandler sets the on-disconnected handler
// function for this connection
func (c *gRPCConnection) SetOnDisconnectedHandler(onDisconnectedHandler server.OnDisconnectedHandler) {
	c.onDisconnectedHandler = onDisconnectedHandler
}
