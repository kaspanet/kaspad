package rejects

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
)

// RejectsContext is the interface for the context needed for the HandleRejects flow.
type HandleRejectsContext interface {
}

type handleRejectsFlow struct {
	HandleRejectsContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRejects handles all reject messages coming through incomingRoute.
// This function assumes that incomingRoute will only return MsgReject.
func HandleRejects(context HandleRejectsContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRejectsFlow{
		HandleRejectsContext: context,
		incomingRoute:        incomingRoute,
		outgoingRoute:        outgoingRoute,
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
