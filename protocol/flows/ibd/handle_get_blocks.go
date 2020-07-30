package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

const ibdBatchSize = 80

// GetBlocksContext is the interface for the context needed for the HandleGetBlocks flow.
type GetBlocksContext interface {
	DAG() *blockdag.BlockDAG
}

type handleGetBlocksFlow struct {
	GetBlocksContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleGetBlocks handles getBlocks messages
func HandleGetBlocks(context GetBlocksContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleGetBlocksFlow{
		GetBlocksContext: context,
		incomingRoute:    incomingRoute,
		outgoingRoute:    outgoingRoute,
	}
	return flow.start()
}

func (flow *handleGetBlocksFlow) start() error {
	for {
		lowHash, highHash, err := receiveGetBlocks(flow.incomingRoute)
		if err != nil {
			return err
		}

		msgIBDBlocks, err := flow.buildMsgIBDBlocks(lowHash, highHash)
		if err != nil {
			return err
		}

		for offset := 0; offset < len(msgIBDBlocks); offset += ibdBatchSize {
			end := offset + ibdBatchSize
			if end > len(msgIBDBlocks) {
				end = len(msgIBDBlocks)
			}

			err = flow.sendMsgIBDBlocks(msgIBDBlocks[offset:end])
			if err != nil {
				return nil
			}

			message, err := flow.incomingRoute.Dequeue()
			if err != nil {
				return err
			}

			if end > len(msgIBDBlocks) {
				break
			}

			if _, ok := message.(*wire.MsgGetNextIBDBlocks); !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", wire.CmdGetNextIBDBlocks, message.Command())
			}
		}
		err = flow.outgoingRoute.Enqueue(wire.NewMsgDoneIBDBlocks())
		if err != nil {
			return err
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

func (flow *handleGetBlocksFlow) buildMsgIBDBlocks(lowHash *daghash.Hash,
	highHash *daghash.Hash) ([]*wire.MsgIBDBlock, error) {

	const maxHashesInMsgIBDBlocks = wire.MaxInvPerMsg
	blockHashes, err := flow.DAG().AntiPastHashesBetween(lowHash, highHash, maxHashesInMsgIBDBlocks)
	if err != nil {
		return nil, err
	}

	msgIBDBlocks := make([]*wire.MsgIBDBlock, len(blockHashes))
	for i, blockHash := range blockHashes {
		block, err := flow.DAG().BlockByHash(blockHash)
		if err != nil {
			return nil, err
		}
		msgIBDBlocks[i] = wire.NewMsgIBDBlock(block.MsgBlock())
	}

	return msgIBDBlocks, nil
}

func (flow *handleGetBlocksFlow) sendMsgIBDBlocks(msgIBDBlocks []*wire.MsgIBDBlock) error {
	for _, msgIBDBlock := range msgIBDBlocks {
		err := flow.outgoingRoute.Enqueue(msgIBDBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
