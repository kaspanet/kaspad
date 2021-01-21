package blockrelay

import (
	"errors"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleRequestPruningPointUTXOSetAndBlockContext is the interface for the context needed for the HandleRequestPruningPointUTXOSetAndBlock flow.
type HandleRequestPruningPointUTXOSetAndBlockContext interface {
	Domain() domain.Domain
}

type handleRequestPruningPointUTXOSetAndBlockFlow struct {
	HandleRequestPruningPointUTXOSetAndBlockContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestPruningPointUTXOSetAndBlock listens to appmessage.MsgRequestPruningPointUTXOSetAndBlock messages and sends
// the pruning point UTXO set and block body.
func HandleRequestPruningPointUTXOSetAndBlock(context HandleRequestPruningPointUTXOSetAndBlockContext, incomingRoute,
	outgoingRoute *router.Route) error {
	flow := &handleRequestPruningPointUTXOSetAndBlockFlow{
		HandleRequestPruningPointUTXOSetAndBlockContext: context,
		incomingRoute: incomingRoute,
		outgoingRoute: outgoingRoute,
	}

	return flow.start()
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) start() error {
	for {
		msgRequestPruningPointUTXOSetAndBlock, err := flow.waitForRequestPruningPointUTXOSetAndBlockMessages()
		if err != nil {
			return err
		}

		err = flow.handleRequestPruningPointUTXOSetAndBlockMessage(msgRequestPruningPointUTXOSetAndBlock)
		if err != nil {
			return err
		}
	}
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) handleRequestPruningPointUTXOSetAndBlockMessage(
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSetAndBlock) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "handleRequestPruningPointUTXOSetAndBlockFlow")
	defer onEnd()

	log.Debugf("Got request for PruningPointHash UTXOSet and Block")

	err := flow.sendPruningPointBlock(msgRequestPruningPointUTXOSetAndBlock)
	if err != nil {
		return err
	}

	return flow.sendPruningPointUTXOSet(msgRequestPruningPointUTXOSetAndBlock)
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) waitForRequestPruningPointUTXOSetAndBlockMessages() (
	*appmessage.MsgRequestPruningPointUTXOSetAndBlock, error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}
	msgRequestPruningPointUTXOSetAndBlock, ok := message.(*appmessage.MsgRequestPruningPointUTXOSetAndBlock)
	if !ok {
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdRequestPruningPointUTXOSetAndBlock, message.Command())
	}
	return msgRequestPruningPointUTXOSetAndBlock, nil
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) sendPruningPointBlock(
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSetAndBlock) error {

	block, err := flow.Domain().Consensus().GetBlock(msgRequestPruningPointUTXOSetAndBlock.PruningPointHash)
	if err != nil {
		return err
	}
	log.Debugf("Retrieved pruning block %s", msgRequestPruningPointUTXOSetAndBlock.PruningPointHash)

	return flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(block)))
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) sendPruningPointUTXOSet(
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSetAndBlock) error {

	// Send the UTXO set in `step`-sized chunks
	const step = 1000
	offset := 0
	chunksSent := 0
	for {
		pruningPointUTXOs, err := flow.Domain().Consensus().GetPruningPointUTXOs(
			msgRequestPruningPointUTXOSetAndBlock.PruningPointHash, offset, step)
		if err != nil {
			if errors.Is(err, ruleerrors.ErrWrongPruningPointHash) {
				return flow.outgoingRoute.Enqueue(appmessage.NewMsgUnexpectedPruningPoint())
			}
		}

		log.Debugf("Retrieved %d UTXOs for pruning block %s",
			len(pruningPointUTXOs), msgRequestPruningPointUTXOSetAndBlock.PruningPointHash)

		outpointAndUTXOEntryPairs :=
			appmessage.DomainOutpointAndUTXOEntryPairsToOutpointAndUTXOEntryPairs(pruningPointUTXOs)
		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgPruningPointUTXOSetChunk(outpointAndUTXOEntryPairs))
		if err != nil {
			return err
		}

		if len(pruningPointUTXOs) < step {
			log.Debugf("Finished sending UTXOs for pruning block %s",
				msgRequestPruningPointUTXOSetAndBlock.PruningPointHash)

			return flow.outgoingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
		}

		offset += step
		chunksSent++

		// Wait for the peer to request more chunks every `ibdBatchSize` chunks
		if chunksSent%ibdBatchSize == 0 {
			message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
			if err != nil {
				return err
			}
			_, ok := message.(*appmessage.MsgRequestNextPruningPointUTXOSetChunk)
			if !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextPruningPointUTXOSetChunk, message.Command())
			}
		}
	}
}
