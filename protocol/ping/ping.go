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

func HandlePing(incomingRoute *router.Route, outgoingRoute *router.Route) error {
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

func StartPingLoop(incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		<-ticker.C

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

		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		pongMessage := message.(*wire.MsgPing)
		if pongMessage.Nonce != pingMessage.Nonce {
			return errors.New("nonce mismatch between ping and pong")
		}
		peer.SetPingIdle()
	}
}
