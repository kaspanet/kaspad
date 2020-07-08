package grpcserver

import "github.com/kaspanet/kaspad/wire"

type gRPCConnection struct{}

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
