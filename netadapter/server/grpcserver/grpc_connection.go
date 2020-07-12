package grpcserver

import (
	"github.com/kaspanet/kaspad/netadapter/router"
	"net"
	"sync"
	"sync/atomic"

	"github.com/kaspanet/kaspad/netadapter/server"
	"google.golang.org/grpc"
)

type gRPCConnection struct {
	server  *gRPCServer
	address net.Addr
	router  *router.Router

	writeDuringDisconnectLock sync.Mutex // writeDuringDisconnectLock makes sure channels aren't written to after close
	errChan                   chan error
	clientConn                grpc.ClientConn
	onDisconnectedHandler     server.OnDisconnectedHandler

	isConnected uint32
}

func newConnection(server *gRPCServer, address net.Addr) *gRPCConnection {
	connection := &gRPCConnection{
		server:      server,
		address:     address,
		errChan:     make(chan error),
		isConnected: 1,
	}

	return connection
}

func (c *gRPCConnection) SetRouter(router *router.Router) {
	c.router = router
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

// Disconnect disconnects the connection
// Calling this function a second time doesn't do anything
//
// This is part of the Connection interface
func (c *gRPCConnection) Disconnect() error {
	if !c.IsConnected() {
		return nil
	}
	atomic.StoreUint32(&c.isConnected, 0)

	c.writeDuringDisconnectLock.Lock()
	defer c.writeDuringDisconnectLock.Unlock()
	close(c.errChan)

	return c.onDisconnectedHandler()
}

func (c *gRPCConnection) Address() net.Addr {
	return c.address
}
