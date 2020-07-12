package grpcserver

import (
	"net"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/netadapter/server"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"
	"github.com/kaspanet/kaspad/wire"
	"google.golang.org/grpc"
)

type gRPCConnection struct {
	server                *gRPCServer
	address               net.Addr
	sendChan              chan *protowire.KaspadMessage
	receiveChan           chan *protowire.KaspadMessage
	errChan               chan error
	clientConn            grpc.ClientConn
	onDisconnectedHandler server.OnDisconnectedHandler

	isConnected uint32
}

func newConnection(server *gRPCServer, address net.Addr) *gRPCConnection {
	connection := &gRPCConnection{
		server:      server,
		address:     address,
		sendChan:    make(chan *protowire.KaspadMessage),
		receiveChan: make(chan *protowire.KaspadMessage),
		errChan:     make(chan error),
		isConnected: 1,
	}

	return connection
}

func (c *gRPCConnection) String() string {
	return c.Address().String()
}

func (c *gRPCConnection) IsConnected() bool {
	return atomic.LoadUint32(&c.isConnected) != 0
}

func (c *gRPCConnection) SetOnDisconnectedHandler(onDisconnectedHandler server.OnDisconnectedHandler) {
	c.onDisconnectedHandler = onDisconnectedHandler
}

// Send sends the given message through the connection
// This is part of the Connection interface
func (c *gRPCConnection) Send(message wire.Message) error {
	messageProto, err := protowire.FromWireMessage(message)
	if err != nil {
		return err
	}

	c.sendChan <- messageProto

	return <-c.errChan
}

// Receive receives the next message from the connection
// This is part of the Connection interface
func (c *gRPCConnection) Receive() (wire.Message, error) {
	protoMessage := <-c.receiveChan
	if protoMessage == nil {
		return nil, errors.New("connection closed during receive")
	}

	return protoMessage.ToWireMessage()
}

// Disconnect disconnects the connection
// This is part of the Connection interface
func (c *gRPCConnection) Disconnect() error {
	if !c.IsConnected() {
		return nil
	}
	atomic.StoreUint32(&c.isConnected, 0)

	close(c.receiveChan)
	close(c.sendChan)
	close(c.errChan)

	return c.onDisconnectedHandler()
}

func (c *gRPCConnection) Address() net.Addr {
	return c.address
}
