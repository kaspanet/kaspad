package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// HandleGetBlockLocator handles getBlockLocator messages
func HandleGetBlockLocator(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		lowHash, highHash, err := receiveGetBlockLocator(incomingRoute)
		if err != nil {
			return err
		}

		locator, err := dag.BlockLocatorFromHashes(highHash, lowHash)
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

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, nil, errors.WithStack(common.ErrRouteClosed)
	}
	msgGetBlockLocator := message.(*wire.MsgGetBlockLocator)

	return msgGetBlockLocator.LowHash, msgGetBlockLocator.HighHash, nil
}

func sendBlockLocator(outgoingRoute *router.Route, locator blockdag.BlockLocator) error {
	msgBlockLocator := wire.NewMsgBlockLocator(locator)
	isOpen := outgoingRoute.Enqueue(msgBlockLocator)
	if !isOpen {
		return errors.WithStack(common.ErrRouteClosed)
	}
	return nil
}
