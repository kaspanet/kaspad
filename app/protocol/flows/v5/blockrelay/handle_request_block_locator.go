package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// RequestBlockLocatorContext is the interface for the context needed for the HandleRequestBlockLocator flow.
type RequestBlockLocatorContext interface {
	Domain() domain.Domain
}

type handleRequestBlockLocatorFlow struct {
	RequestBlockLocatorContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestBlockLocator handles getBlockLocator messages
func HandleRequestBlockLocator(context RequestBlockLocatorContext, incomingRoute *router.Route,
	outgoingRoute *router.Route) error {

	flow := &handleRequestBlockLocatorFlow{
		RequestBlockLocatorContext: context,
		incomingRoute:              incomingRoute,
		outgoingRoute:              outgoingRoute,
	}
	return flow.start()
}

func (flow *handleRequestBlockLocatorFlow) start() error {
	for {
		highHash, limit, err := flow.receiveGetBlockLocator()
		if err != nil {
			return err
		}
		log.Tracef("Received getBlockLocator with highHash: %s, limit: %d", highHash, limit)

		locator, err := flow.Domain().Consensus().CreateBlockLocatorFromPruningPoint(highHash, limit)
		if err != nil || len(locator) == 0 {
			if err != nil {
				log.Tracef("Received error from CreateBlockLocatorFromPruningPoint: %s", err)
			}
			return protocolerrors.Errorf(true, "couldn't build a block "+
				"locator between the pruning point and %s", highHash)
		}

		err = flow.sendBlockLocator(locator)
		if err != nil {
			return err
		}
	}
}

func (flow *handleRequestBlockLocatorFlow) receiveGetBlockLocator() (highHash *externalapi.DomainHash, limit uint32, err error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, 0, err
	}
	msgGetBlockLocator := message.(*appmessage.MsgRequestBlockLocator)

	return msgGetBlockLocator.HighHash, msgGetBlockLocator.Limit, nil
}

func (flow *handleRequestBlockLocatorFlow) sendBlockLocator(locator externalapi.BlockLocator) error {
	msgBlockLocator := appmessage.NewMsgBlockLocator(locator)
	err := flow.outgoingRoute.Enqueue(msgBlockLocator)
	if err != nil {
		return err
	}
	return nil
}
