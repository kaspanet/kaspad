package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandlePruningPointHashRequestsFlowContext is the interface for the context needed for the handlePruningPointHashRequestsFlow flow.
type HandlePruningPointHashRequestsFlowContext interface {
	Domain() domain.Domain
}

type handlePruningPointHashRequestsFlow struct {
	HandlePruningPointHashRequestsFlowContext
	incomingRoute, outgoingRoute *router.Route
}

// HandlePruningPointHashRequests listens to appmessage.MsgRequestPruningPointHashMessage messages and sends
// the pruning point hash as response.
func HandlePruningPointHashRequests(context HandlePruningPointHashRequestsFlowContext, incomingRoute,
	outgoingRoute *router.Route) error {
	flow := &handlePruningPointHashRequestsFlow{
		HandlePruningPointHashRequestsFlowContext: context,
		incomingRoute: incomingRoute,
		outgoingRoute: outgoingRoute,
	}

	return flow.start()
}

func (flow *handlePruningPointHashRequestsFlow) start() error {
	for {
		_, err := flow.incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		log.Debugf("Got request for a pruning point hash")

		pruningPoint, err := flow.Domain().Consensus().PruningPoint()
		if err != nil {
			return err
		}

		err = flow.outgoingRoute.Enqueue(appmessage.NewPruningPointHashMessage(pruningPoint))
		if err != nil {
			return err
		}
		log.Debugf("Sent pruning point hash %s", pruningPoint)
	}
}
