package server

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/netadapter/router"
)

// OnConnectedHandler is a function that is to be called
// once a new Connection is successfully established.
type OnConnectedHandler func(connection Connection) error

// OnDisconnectedHandler is a function that is to be
// called once a Connection has been disconnected.
type OnDisconnectedHandler func() error

// Server represents a p2p server.
type Server interface {
	Connect(address string) (Connection, error)
	Start() error
	Stop() error
	SetOnConnectedHandler(onConnectedHandler OnConnectedHandler)
	OnConnectedHandler() OnConnectedHandler
}

// Connection represents a p2p server connection.
type Connection interface {
	fmt.Stringer
	Start(router *router.Router)
	Disconnect() error
	IsConnected() bool
	SetOnDisconnectedHandler(onDisconnectedHandler OnDisconnectedHandler)
	Address() net.Addr
}

// ErrNetwork is an error related to the internals of the connection, and not an error that
// came from outside (e.g. from OnDisconnectedHandler).
var ErrNetwork = errors.New("network error")
