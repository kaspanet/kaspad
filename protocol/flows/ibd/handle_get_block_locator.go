package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// HandleGetBlockLocator handles getBlockLocator messages
func HandleGetBlockLocator(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		lowHash, highHash, shouldStop, err := receiveGetBlockLocator(incomingRoute)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}

		locator, err := dag.BlockLocatorFromHashes(highHash, lowHash)
		if err != nil || len(locator) == 0 {
			return protocolerrors.Errorf(true, "couldn't build a block "+
				"locator between blocks %s and %s", lowHash, highHash)
		}

		shouldStop, err = sendBlockLocator(outgoingRoute, locator)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
}

func receiveGetBlockLocator(incomingRoute *router.Route) (lowHash *daghash.Hash,
	highHash *daghash.Hash, shouldStop bool, err error) {

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, nil, true, nil
	}
	msgGetBlockLocator := message.(*wire.MsgGetBlockLocator)

	return msgGetBlockLocator.LowHash, msgGetBlockLocator.HighHash, false, nil
}

func sendBlockLocator(outgoingRoute *router.Route, locator blockdag.BlockLocator) (shouldStop bool, err error) {
	msgBlockLocator := wire.NewMsgBlockLocator(locator)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgBlockLocator, common.DefaultTimeout)
	if err != nil {
		return true, err
	}
	return !isOpen, nil
}
