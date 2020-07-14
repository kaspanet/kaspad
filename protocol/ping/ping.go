package ping

import (
	"errors"
	"github.com/kaspanet/kaspad/netadapter/router"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/random"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

const pingInterval = 2 * time.Minute

// HandlePing handles all ping messages coming through incomingRoute.
// This function assumes that incomingRoute will only return MsgPing.
func HandlePing(incomingRoute *router.Route, outgoingRoute *router.Route) error {
	for {
		message, isOpen := incomingRoute.Dequeue()
		if !isOpen {
			return nil
		}
		pingMessage := message.(*wire.MsgPing)

		pongMessage := wire.NewMsgPong(pingMessage.Nonce)
		isOpen = outgoingRoute.Enqueue(pongMessage)
		if !isOpen {
			return nil
		}
	}
}

// StartPingLoop starts the ping-sending loop, which sends MsgPing to
// the given peer every pingInterval.
// This function assumes that incomingRoute will only return MsgPong.
func StartPingLoop(incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		nonce, err := random.Uint64()
		if err != nil {
			return err
		}
		peer.SetPingPending(nonce)

		pingMessage := wire.NewMsgPing(nonce)
		isOpen := outgoingRoute.Enqueue(pingMessage)
		if !isOpen {
			return nil
		}

		message, isOpen := incomingRoute.Dequeue()
		if !isOpen {
			return nil
		}
		pongMessage := message.(*wire.MsgPing)
		if pongMessage.Nonce != pingMessage.Nonce {
			return errors.New("nonce mismatch between ping and pong")
		}
		peer.SetPingIdle()
	}
}
