package ibd

import (
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/kaspanet/kaspad/network/netadapter/router"
	"github.com/kaspanet/kaspad/network/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
)

// RequestBlockLocatorContext is the interface for the context needed for the HandleRequestBlockLocator flow.
type RequestBlockLocatorContext interface {
	DAG() *blockdag.BlockDAG
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
		lowHash, highHash, err := flow.receiveGetBlockLocator()
		if err != nil {
			return err
		}

		locator, err := flow.DAG().BlockLocatorFromHashes(highHash, lowHash)
		if err != nil || len(locator) == 0 {
			return protocolerrors.Errorf(true, "couldn't build a block "+
				"locator between blocks %s and %s", lowHash, highHash)
		}

		err = flow.sendBlockLocator(locator)
		if err != nil {
			return err
		}
	}
}

func (flow *handleRequestBlockLocatorFlow) receiveGetBlockLocator() (lowHash *daghash.Hash,
	highHash *daghash.Hash, err error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgGetBlockLocator := message.(*appmessage.MsgRequestBlockLocator)

	return msgGetBlockLocator.LowHash, msgGetBlockLocator.HighHash, nil
}

func (flow *handleRequestBlockLocatorFlow) sendBlockLocator(locator blockdag.BlockLocator) error {
	msgBlockLocator := appmessage.NewMsgBlockLocator(locator)
	err := flow.outgoingRoute.Enqueue(msgBlockLocator)
	if err != nil {
		return err
	}
	return nil
}
