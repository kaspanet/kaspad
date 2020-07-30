package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// GetBlockLocatorContext is the interface for the context needed for the HandleGetBlockLocator flow.
type GetBlockLocatorContext interface {
	DAG() *blockdag.BlockDAG
}

type handleGetBlockLocatorFlow struct {
	GetBlockLocatorContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleGetBlockLocator handles getBlockLocator messages
func HandleGetBlockLocator(context GetBlockLocatorContext, incomingRoute *router.Route,
	outgoingRoute *router.Route) error {

	flow := &handleGetBlockLocatorFlow{
		GetBlockLocatorContext: context,
		incomingRoute:          incomingRoute,
		outgoingRoute:          outgoingRoute,
	}
	return flow.start()
}

func (flow *handleGetBlockLocatorFlow) start() error {
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

func (flow *handleGetBlockLocatorFlow) receiveGetBlockLocator() (lowHash *daghash.Hash,
	highHash *daghash.Hash, err error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgGetBlockLocator := message.(*wire.MsgRequestBlockLocator)

	return msgGetBlockLocator.LowHash, msgGetBlockLocator.HighHash, nil
}

func (flow *handleGetBlockLocatorFlow) sendBlockLocator(locator blockdag.BlockLocator) error {
	msgBlockLocator := wire.NewMsgBlockLocator(locator)
	err := flow.outgoingRoute.Enqueue(msgBlockLocator)
	if err != nil {
		return err
	}
	return nil
}
