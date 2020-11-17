package standalone

import (
	"time"

	"github.com/kaspanet/kaspad/infrastructure/network/netadapter"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// Routes holds the incoming and outgoing routes of a connection created by MinimalNetAdapter
type Routes struct {
	netConnection                *netadapter.NetConnection
	IncomingRoute, OutgoingRoute *router.Route
	handshakeRoute               *router.Route
	addressesRoute               *router.Route
	pingRoute                    *router.Route
}

// WaitForMessageOfType waits for a message of requested type up to `timeout`, skipping all messages of any other type
// received while waiting
func (r *Routes) WaitForMessageOfType(command appmessage.MessageCommand, timeout time.Duration) (appmessage.Message, error) {
	timeoutTime := time.Now().Add(timeout)
	for {
		route := r.chooseRouteForCommand(command)
		message, err := route.DequeueWithTimeout(timeoutTime.Sub(time.Now()))
		if err != nil {
			return nil, errors.Wrapf(err, "error waiting for message of type %s", command)
		}
		if message.Command() == command {
			return message, nil
		}
	}
}

func (r *Routes) chooseRouteForCommand(command appmessage.MessageCommand) *router.Route {
	switch command {
	case appmessage.CmdVersion, appmessage.CmdVerAck:
		return r.handshakeRoute
	case appmessage.CmdRequestAddresses, appmessage.CmdAddresses:
		return r.addressesRoute
	case appmessage.CmdPing:
		return r.pingRoute
	default:
		return r.IncomingRoute
	}
}

// WaitForDisconnect waits for a disconnect up to `timeout`, skipping all messages received while waiting
func (r *Routes) WaitForDisconnect(timeout time.Duration) error {
	timeoutTime := time.Now().Add(timeout)
	for {
		_, err := r.IncomingRoute.DequeueWithTimeout(timeoutTime.Sub(time.Now()))
		if errors.Is(err, router.ErrRouteClosed) {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "error waiting for disconnect")
		}
	}
}

// Disconnect closes the connection behind the routes, thus closing all routes
func (r *Routes) Disconnect() {
	r.netConnection.Disconnect()
}
