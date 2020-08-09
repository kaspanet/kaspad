package ping

import (
	"time"

	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/random"
	"github.com/kaspanet/kaspad/wire"
)

// SendPingsContext is the interface for the context needed for the SendPings flow.
type SendPingsContext interface {
}

type sendPingsFlow struct {
	SendPingsContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// SendPings starts sending MsgPings every pingInterval seconds to the
// given peer.
// This function assumes that incomingRoute will only return MsgPong.
func SendPings(context SendPingsContext, incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {
	flow := &sendPingsFlow{
		SendPingsContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
		peer:             peer,
	}
	return flow.start()
}

func (flow *sendPingsFlow) start() error {
	const pingInterval = 2 * time.Minute
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		nonce, err := random.Uint64()
		if err != nil {
			return err
		}
		flow.peer.SetPingPending(nonce)

		pingMessage := wire.NewMsgPing(nonce)
		err = flow.outgoingRoute.Enqueue(pingMessage)
		if err != nil {
			return err
		}

		message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
		if err != nil {
			return err
		}
		pongMessage := message.(*wire.MsgPong)
		if pongMessage.Nonce != pingMessage.Nonce {
			return protocolerrors.New(true, "nonce mismatch between ping and pong")
		}
		flow.peer.SetPingIdle()
	}
	return nil
}
