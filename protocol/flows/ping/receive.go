package ping

import (
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
)

// ReceivePingsContext is the interface for the context needed for the ReceivePings flow.
type ReceivePingsContext interface {
}

type receivePingsFlow struct {
	incomingRoute, outgoingRoute *router.Route
}

// ReceivePings handles all ping messages coming through incomingRoute.
// This function assumes that incomingRoute will only return MsgPing.
func ReceivePings(_ ReceivePingsContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &receivePingsFlow{
		incomingRoute: incomingRoute,
		outgoingRoute: outgoingRoute,
	}
	return flow.start()
}

func (flow *receivePingsFlow) start() error {
	for {
		message, err := flow.incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		pingMessage := message.(*wire.MsgPing)

		pongMessage := wire.NewMsgPong(pingMessage.Nonce)
		err = flow.outgoingRoute.Enqueue(pongMessage)
		if err != nil {
			return err
		}
	}
}
