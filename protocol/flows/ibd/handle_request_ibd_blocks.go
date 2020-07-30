package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
)

// RequestIBDBlocksContext is the interface for the context needed for the HandleRequestIBDBlocks flow.
type RequestIBDBlocksContext interface {
	DAG() *blockdag.BlockDAG
}

type handleRequestBlocksFlow struct {
	RequestIBDBlocksContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestIBDBlocks handles getBlocks messages
func HandleRequestIBDBlocks(context RequestIBDBlocksContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRequestBlocksFlow{
		RequestIBDBlocksContext: context,
		incomingRoute:           incomingRoute,
		outgoingRoute:           outgoingRoute,
	}
	return flow.start()
}

func (flow *handleRequestBlocksFlow) start() error {
	for {
		lowHash, highHash, err := receiveRequestIBDBlocks(flow.incomingRoute)
		if err != nil {
			return err
		}

		msgIBDBlocks, err := flow.buildMsgIBDBlocks(lowHash, highHash)
		if err != nil {
			return err
		}

		err = flow.sendMsgIBDBlocks(msgIBDBlocks)
		if err != nil {
			return nil
		}
	}
}

func receiveRequestIBDBlocks(incomingRoute *router.Route) (lowHash *daghash.Hash,
	highHash *daghash.Hash, err error) {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgRequestIBDBlocks := message.(*wire.MsgRequestIBDBlocks)

	return msgRequestIBDBlocks.LowHash, msgRequestIBDBlocks.HighHash, nil
}

func (flow *handleRequestBlocksFlow) buildMsgIBDBlocks(lowHash *daghash.Hash,
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

func (flow *handleRequestBlocksFlow) sendMsgIBDBlocks(msgIBDBlocks []*wire.MsgIBDBlock) error {
	for _, msgIBDBlock := range msgIBDBlocks {
		err := flow.outgoingRoute.Enqueue(msgIBDBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
