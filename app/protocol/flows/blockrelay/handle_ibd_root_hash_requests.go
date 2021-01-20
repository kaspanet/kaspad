package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleIBDRootHashRequestsFlowContext is the interface for the context needed for the handleIBDRootHashRequestsFlow flow.
type HandleIBDRootHashRequestsFlowContext interface {
	Domain() domain.Domain
}

type handleIBDRootHashRequestsFlow struct {
	HandleIBDRootHashRequestsFlowContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleIBDRootHashRequests listens to appmessage.MsgRequestPruningPointHashMessage messages and sends
// the IBD root hash as response.
func HandleIBDRootHashRequests(context HandleIBDRootHashRequestsFlowContext, incomingRoute,
	outgoingRoute *router.Route) error {
	flow := &handleIBDRootHashRequestsFlow{
		HandleIBDRootHashRequestsFlowContext: context,
		incomingRoute:                        incomingRoute,
		outgoingRoute:                        outgoingRoute,
	}

	return flow.start()
}

func (flow *handleIBDRootHashRequestsFlow) start() error {
	for {
		_, err := flow.incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		log.Debugf("Got request for IBD root hash")

		pruningPoint, err := flow.Domain().Consensus().PruningPoint()
		if err != nil {
			return err
		}

		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDRootHashMessage(pruningPoint))
		if err != nil {
			return err
		}
		log.Debugf("Sent IBD root hash %s", pruningPoint)
	}
}
