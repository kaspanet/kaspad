package rejects

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
)

// RejectsContext is the interface for the context needed for the HandleRejects flow.
type RejectsContext interface {
}

type handleRejectsFlow struct {
	RejectsContext
	incomingRoute, outgoingRoute *router.Route
}

// ReceivePings handles all ping messages coming through incomingRoute.
// This function assumes that incomingRoute will only return MsgPing.
func HandleRejects(context RejectsContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRejectsFlow{
		RejectsContext: context,
		incomingRoute:  incomingRoute,
		outgoingRoute:  outgoingRoute,
	}
	return flow.start()
}

func (flow *handleRejectsFlow) start() error {
	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return err
	}
	rejectMessage := message.(*domainmessage.MsgReject)

	const maxReasonLength = 255
	if len(rejectMessage.Reason) > maxReasonLength {
		return protocolerrors.Errorf(false, "got reject message longer than %d", maxReasonLength)
	}

	return protocolerrors.Errorf(false, "got reject message: `%s`", rejectMessage.Reason)
}
