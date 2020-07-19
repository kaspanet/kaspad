package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// HandleGetBlocks handles getBlocks messages
func HandleGetBlocks(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		lowHash, highHash, shouldStop, err := receiveGetBlocks(incomingRoute)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}

		msgIBDBlocks, err := buildMsgIBDBlocks(lowHash, highHash, dag)
		if err != nil {
			return err
		}

		shouldStop, err = sendMsgIBDBlocks(outgoingRoute, msgIBDBlocks)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
}

func receiveGetBlocks(incomingRoute *router.Route) (lowHash *daghash.Hash,
	highHash *daghash.Hash, shouldStop bool, err error) {

	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, nil, true, nil
	}
	msgGetBlocks := message.(*wire.MsgGetBlocks)

	return msgGetBlocks.LowHash, msgGetBlocks.HighHash, false, nil
}

func buildMsgIBDBlocks(lowHash *daghash.Hash, highHash *daghash.Hash,
	dag *blockdag.BlockDAG) ([]*wire.MsgIBDBlock, error) {

	blockHashes, err := dag.AntiPastHashesBetween(lowHash, highHash, wire.MaxInvPerMsg)
	if err != nil {
		return nil, err
	}

	msgIBDBlocks := make([]*wire.MsgIBDBlock, len(blockHashes))
	for i, blockHash := range blockHashes {
		block, err := dag.BlockByHash(blockHash)
		if err != nil {
			return nil, err
		}
		msgIBDBlocks[i] = wire.NewMsgIBDBlock(block.MsgBlock())
	}

	return msgIBDBlocks, nil
}

func sendMsgIBDBlocks(outgoingRoute *router.Route, msgIBDBlocks []*wire.MsgIBDBlock) (shouldStop bool, err error) {
	for _, msgIBDBlock := range msgIBDBlocks {
		isOpen, err := outgoingRoute.EnqueueWithTimeout(msgIBDBlock, common.DefaultTimeout)
		if err != nil {
			return true, err
		}
		if !isOpen {
			return true, nil
		}
	}
	return false, nil
}
