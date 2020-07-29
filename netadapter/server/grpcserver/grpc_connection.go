package grpcserver

import (
	"net"
	"sync/atomic"

	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/netadapter/server/grpcserver/protowire"

	"github.com/kaspanet/kaspad/netadapter/server"
	"google.golang.org/grpc"
)

type gRPCConnection struct {
	server     *gRPCServer
	address    *net.TCPAddr
	isOutbound bool
	stream     grpcStream
	router     *router.Router

	stopChan                chan struct{}
	clientConn              grpc.ClientConn
	onDisconnectedHandler   server.OnDisconnectedHandler
	onInvalidMessageHandler server.OnInvalidMessageHandler

	isConnected uint32
}

func newConnection(server *gRPCServer, address *net.TCPAddr, isOutbound bool, stream grpcStream) *gRPCConnection {
	connection := &gRPCConnection{
		server:      server,
		address:     address,
		isOutbound:  isOutbound,
		stream:      stream,
		stopChan:    make(chan struct{}),
		isConnected: 1,
	}

	return connection
}

func (c *gRPCConnection) Start(router *router.Router) {
	c.router = router

	spawn("gRPCConnection.Start-connectionLoops", func() {
		err := c.connectionLoops()
		if err != nil {
			log.Errorf("error from connectionLoops for %s: %+v", c.address, err)
		}
	})
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

func (c *gRPCConnection) SetOnInvalidMessageHandler(onInvalidMessageHandler server.OnInvalidMessageHandler) {
	c.onInvalidMessageHandler = onInvalidMessageHandler
}

func (c *gRPCConnection) IsOutbound() bool {
	return c.isOutbound
}

// Disconnect disconnects the connection
// Calling this function a second time doesn't do anything
//
// This is part of the Connection interface
func (c *gRPCConnection) Disconnect() {
	if !c.IsConnected() {
		return
	}
	atomic.StoreUint32(&c.isConnected, 0)

	close(c.stopChan)

	if c.isOutbound {
		clientStream := c.stream.(protowire.P2P_MessageStreamClient)
		_ = clientStream.CloseSend() // ignore error because we don't really know what's the status of the connection
	}

	log.Debugf("Disconnected from %s", c)

	if c.onDisconnectedHandler != nil {
		c.onDisconnectedHandler()
	}
}

func (c *gRPCConnection) Address() *net.TCPAddr {
	return c.address
}
