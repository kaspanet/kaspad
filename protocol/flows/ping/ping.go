package ping

import (
	"github.com/kaspanet/kaspad/protocol/common"
	"time"

	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/random"
	"github.com/kaspanet/kaspad/wire"
)

// ReceivePingsContext is the interface for the context needed for the ReceivePings flow.
type ReceivePingsContext interface {
}

// ReceivePings handles all ping messages coming through incomingRoute.
// This function assumes that incomingRoute will only return MsgPing.
func ReceivePings(_ ReceivePingsContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		pingMessage := message.(*wire.MsgPing)

		pongMessage := wire.NewMsgPong(pingMessage.Nonce)
		err = outgoingRoute.Enqueue(pongMessage)
		if err != nil {
			return err
		}
	}
}

// SendPingsContext is the interface for the context needed for the SendPings flow.
type SendPingsContext interface {
}

// SendPings starts sending MsgPings every pingInterval seconds to the
// given peer.
// This function assumes that incomingRoute will only return MsgPong.
func SendPings(_ SendPingsContext, incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {
	const pingInterval = 2 * time.Minute
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		nonce, err := random.Uint64()
		if err != nil {
			return err
		}
		peer.SetPingPending(nonce)

		pingMessage := wire.NewMsgPing(nonce)
		err = outgoingRoute.Enqueue(pingMessage)
		if err != nil {
			return err
		}

		message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return err
		}
		pongMessage := message.(*wire.MsgPong)
		if pongMessage.Nonce != pingMessage.Nonce {
			return protocolerrors.New(true, "nonce mismatch between ping and pong")
		}
		peer.SetPingIdle()
	}
	return nil
}
