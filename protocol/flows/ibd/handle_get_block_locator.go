package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

type GetBlockLocatorContext interface {
	DAG() *blockdag.BlockDAG
}

// HandleGetBlockLocator handles getBlockLocator messages
func HandleGetBlockLocator(context GetBlockLocatorContext, incomingRoute *router.Route,
	outgoingRoute *router.Route) error {

	for {
		lowHash, highHash, err := receiveGetBlockLocator(incomingRoute)
		if err != nil {
			return err
		}

		locator, err := context.DAG().BlockLocatorFromHashes(highHash, lowHash)
		if err != nil || len(locator) == 0 {
			return protocolerrors.Errorf(true, "couldn't build a block "+
				"locator between blocks %s and %s", lowHash, highHash)
		}

		err = sendBlockLocator(outgoingRoute, locator)
		if err != nil {
			return err
		}
	}
}

func receiveGetBlockLocator(incomingRoute *router.Route) (lowHash *daghash.Hash,
	highHash *daghash.Hash, err error) {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgGetBlockLocator := message.(*wire.MsgGetBlockLocator)

	return msgGetBlockLocator.LowHash, msgGetBlockLocator.HighHash, nil
}

func sendBlockLocator(outgoingRoute *router.Route, locator blockdag.BlockLocator) error {
	msgBlockLocator := wire.NewMsgBlockLocator(locator)
	err := outgoingRoute.Enqueue(msgBlockLocator)
	if err != nil {
		return err
	}
	return nil
}
