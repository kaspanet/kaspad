package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

func HandleGetBlockLocator(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		lowHash, highHash, shouldContinue, err := receiveGetBlockLocator(incomingRoute)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}

		locator, err := dag.BlockLocatorFromHashes(highHash, lowHash)
		if err != nil || len(locator) == 0 {
			return protocolerrors.Errorf(true, "couldn't build a block "+
				"locator between blocks %s and %s", lowHash, highHash)
		}

		shouldContinue, err = sendBlockLocator(outgoingRoute, locator)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}
	}
}

func receiveGetBlockLocator(incomingRoute *router.Route) (lowHash *daghash.Hash,
	highHash *daghash.Hash, shouldContinue bool, err error) {

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, nil, false, nil
	}
	msgGetBlockLocator := message.(*wire.MsgGetBlockLocator)

	return msgGetBlockLocator.LowHash, msgGetBlockLocator.HighHash, true, nil
}

func sendBlockLocator(outgoingRoute *router.Route, locator blockdag.BlockLocator) (shouldContinue bool, err error) {
	msgBlockLocator := wire.NewMsgBlockLocator()
	for _, hash := range locator {
		err := msgBlockLocator.AddBlockLocatorHash(hash)
		if err != nil {
			return false, err
		}
	}

	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgBlockLocator, common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	return isOpen, nil

}
