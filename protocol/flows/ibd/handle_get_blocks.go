package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

type GetBlocksContext interface {
	DAG() *blockdag.BlockDAG
}

// HandleGetBlocks handles getBlocks messages
func HandleGetBlocks(context GetBlocksContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	for {
		lowHash, highHash, err := receiveGetBlocks(incomingRoute)
		if err != nil {
			return err
		}

		msgIBDBlocks, err := buildMsgIBDBlocks(context, lowHash, highHash)
		if err != nil {
			return err
		}

		err = sendMsgIBDBlocks(outgoingRoute, msgIBDBlocks)
		if err != nil {
			return nil
		}
	}
}

func receiveGetBlocks(incomingRoute *router.Route) (lowHash *daghash.Hash,
	highHash *daghash.Hash, err error) {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgGetBlocks := message.(*wire.MsgGetBlocks)

	return msgGetBlocks.LowHash, msgGetBlocks.HighHash, nil
}

func buildMsgIBDBlocks(context GetBlocksContext, lowHash *daghash.Hash,
	highHash *daghash.Hash) ([]*wire.MsgIBDBlock, error) {

	const maxHashesInMsgIBDBlocks = wire.MaxInvPerMsg
	blockHashes, err := context.DAG().AntiPastHashesBetween(lowHash, highHash, maxHashesInMsgIBDBlocks)
	if err != nil {
		return nil, err
	}

	msgIBDBlocks := make([]*wire.MsgIBDBlock, len(blockHashes))
	for i, blockHash := range blockHashes {
		block, err := context.DAG().BlockByHash(blockHash)
		if err != nil {
			return nil, err
		}
		msgIBDBlocks[i] = wire.NewMsgIBDBlock(block.MsgBlock())
	}

	return msgIBDBlocks, nil
}

func sendMsgIBDBlocks(outgoingRoute *router.Route, msgIBDBlocks []*wire.MsgIBDBlock) error {
	for _, msgIBDBlock := range msgIBDBlocks {
		err := outgoingRoute.Enqueue(msgIBDBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
