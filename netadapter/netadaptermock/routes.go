package netadaptermock

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
)

// Routes holds the incoming and outgoing routes of a connection created by NetAdapterMock
type Routes struct {
	IncomingRoute, OutgoingRoute *router.Route
	handshakeRoute               *router.Route
	pingRoute                    *router.Route
}

// WaitForMessageOfType waits for a message of requested type up to `timeout`, skipping all messages of any other type
// received while waiting
func (r *Routes) WaitForMessageOfType(command wire.MessageCommand, timeout time.Duration) (wire.Message, error) {
	timeoutTime := time.Now().Add(timeout)
	for {
		message, err := r.IncomingRoute.DequeueWithTimeout(timeoutTime.Sub(time.Now()))
		if err != nil {
			return nil, errors.Wrapf(err, "Error waiting for message of type %s", command)
		}
		if message.Command() == command {
			return message, nil
		}
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
			return errors.Wrap(err, "Error waiting for disconnect")
		}
	}
}

// Close closes all the routes in this Routes object
func (r *Routes) Close() {
	r.IncomingRoute.Close()
	r.OutgoingRoute.Close()
	r.handshakeRoute.Close()
	r.pingRoute.Close()
}
