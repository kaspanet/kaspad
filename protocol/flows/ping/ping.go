package ping

import (
	"time"

	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/random"
	"github.com/kaspanet/kaspad/wire"
)

const pingTimeout = 30 * time.Second

// ReceivePings handles all ping messages coming through incomingRoute.
// This function assumes that incomingRoute will only return MsgPing.
func ReceivePings(incomingRoute *router.Route, outgoingRoute *router.Route) error {
	for {
		message, isOpen := incomingRoute.Dequeue()
		if !isOpen {
			return nil
		}
		pingMessage := message.(*wire.MsgPing)

		pongMessage := wire.NewMsgPong(pingMessage.Nonce)
		isOpen, err := outgoingRoute.EnqueueWithTimeout(pongMessage, pingTimeout)
		if err != nil {
			return err
		}
		if !isOpen {
			return nil
		}
	}
}

// SendPings starts sending MsgPings every pingInterval seconds to the
// given peer.
// This function assumes that incomingRoute will only return MsgPong.
func SendPings(incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {
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
		isOpen, err := outgoingRoute.EnqueueWithTimeout(pingMessage, pingTimeout)
		if err != nil {
			return err
		}
		if !isOpen {
			return nil
		}

		message, isOpen, err := incomingRoute.DequeueWithTimeout(pingTimeout)
		if err != nil {
			return err
		}
		if !isOpen {
			return nil
		}
		pongMessage := message.(*wire.MsgPong)
		if pongMessage.Nonce != pingMessage.Nonce {
			return protocolerrors.New(true, "nonce mismatch between ping and pong")
		}
		peer.SetPingIdle()
	}
	return nil
}
