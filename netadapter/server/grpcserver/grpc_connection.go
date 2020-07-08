package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/wire"
)

type gRPCConnection struct {
	isClient     bool
	clientStream protowire.P2P_MessageStreamClient
	serverStream protowire.P2P_MessageStreamServer
}

func newClientConnection(stream protowire.P2P_MessageStreamClient) *gRPCConnection {
	return &gRPCConnection{
		isClient:     true,
		clientStream: stream,
	}
}

func newServerConnection(stream protowire.P2P_MessageStreamServer) *gRPCConnection {
	return &gRPCConnection{
		isClient:     false,
		serverStream: stream,
	}
}

// Send sends the given message through the connection
// This is part of the Connection interface
func (c *gRPCConnection) Send(message wire.Message) error {
	messageProto, err := protowire.FromWireMessage(message)
	if err != nil {
		return err
	}

	if c.isClient {
		return c.clientStream.Send(messageProto)
	}
	return c.serverStream.Send(messageProto)
}

// Receive receives the next message from the connection
// This is part of the Connection interface
func (c *gRPCConnection) Receive() (wire.Message, error) {
	var protoMessage *protowire.KaspadMessage
	var err error

	if c.isClient {
		protoMessage, err = c.clientStream.Recv()
	} else {
		protoMessage, err = c.clientStream.Recv()
	}
	if err != nil {
		return nil, err
	}

	return protoMessage.ToWireMessage()
}

// Disconnect disconnects the connection
// This is part of the Connection interface
func (c *gRPCConnection) Disconnect() error {
	// TODO(libp2p): unimplemented
	panic("unimplemented")
}
