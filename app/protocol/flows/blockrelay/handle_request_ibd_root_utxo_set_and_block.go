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

// HandleRequestIBDRootUTXOSetAndBlockContext is the interface for the context needed for the HandleRequestIBDRootUTXOSetAndBlock flow.
type HandleRequestIBDRootUTXOSetAndBlockContext interface {
	Domain() domain.Domain
}

type handleRequestIBDRootUTXOSetAndBlockFlow struct {
	HandleRequestIBDRootUTXOSetAndBlockContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestIBDRootUTXOSetAndBlock listens to appmessage.MsgRequestIBDRootUTXOSetAndBlock messages and sends
// the IBD root UTXO set and block body.
func HandleRequestIBDRootUTXOSetAndBlock(context HandleRequestIBDRootUTXOSetAndBlockContext, incomingRoute,
	outgoingRoute *router.Route) error {
	flow := &handleRequestIBDRootUTXOSetAndBlockFlow{
		HandleRequestIBDRootUTXOSetAndBlockContext: context,
		incomingRoute: incomingRoute,
		outgoingRoute: outgoingRoute,
	}

	return flow.start()
}

func (flow *handleRequestIBDRootUTXOSetAndBlockFlow) start() error {
	for {
		msgRequestIBDRootUTXOSetAndBlock, err := flow.waitForRequestIBDRootUTXOSetAndBlockMessages()
		if err != nil {
			return err
		}

		finishMeasuring := logger.LogAndMeasureExecutionTime(log, "handleRequestIBDRootUTXOSetAndBlockFlow")
		log.Debugf("Got request for IBDRoot UTXOSet and Block")

		err = flow.sendIBDRootBlock(msgRequestIBDRootUTXOSetAndBlock)
		if err != nil {
			return err
		}

		err = flow.sendIBDRootUTXOSet(msgRequestIBDRootUTXOSetAndBlock)
		if err != nil {
			return err
		}

		finishMeasuring()
	}
}

func (flow *handleRequestIBDRootUTXOSetAndBlockFlow) waitForRequestIBDRootUTXOSetAndBlockMessages() (
	*appmessage.MsgRequestIBDRootUTXOSetAndBlock, error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}
	msgRequestIBDRootUTXOSetAndBlock, ok := message.(*appmessage.MsgRequestIBDRootUTXOSetAndBlock)
	if !ok {
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdRequestIBDRootUTXOSetAndBlock, message.Command())
	}
	return msgRequestIBDRootUTXOSetAndBlock, nil
}

func (flow *handleRequestIBDRootUTXOSetAndBlockFlow) sendIBDRootBlock(
	msgRequestIBDRootUTXOSetAndBlock *appmessage.MsgRequestIBDRootUTXOSetAndBlock) error {

	block, err := flow.Domain().Consensus().GetBlock(msgRequestIBDRootUTXOSetAndBlock.IBDRoot)
	if err != nil {
		return err
	}
	log.Debugf("Retrieved pruning block %s", msgRequestIBDRootUTXOSetAndBlock.IBDRoot)

	return flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(block)))
}

func (flow *handleRequestIBDRootUTXOSetAndBlockFlow) sendIBDRootUTXOSet(
	msgRequestIBDRootUTXOSetAndBlock *appmessage.MsgRequestIBDRootUTXOSetAndBlock) error {

	// Send the UTXO set in `step`-sized chunks
	const step = 1000
	offset := 0
	chunksSent := 0
	for {
		pruningPointUTXOs, err := flow.Domain().Consensus().GetPruningPointUTXOs(
			msgRequestIBDRootUTXOSetAndBlock.IBDRoot, offset, step)
		if err != nil {
			if errors.Is(err, ruleerrors.ErrWrongPruningPointHash) {
				return flow.outgoingRoute.Enqueue(appmessage.NewMsgUnexpectedPruningPoint())
			}
		}

		log.Debugf("Retrieved %d UTXOs for pruning block %s",
			len(pruningPointUTXOs), msgRequestIBDRootUTXOSetAndBlock.IBDRoot)

		outpointAndUTXOEntryPairs :=
			appmessage.DomainOutpointAndUTXOEntryPairsToOutpointAndUTXOEntryPairs(pruningPointUTXOs)
		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDRootUTXOSetChunk(outpointAndUTXOEntryPairs))
		if err != nil {
			return err
		}

		if len(pruningPointUTXOs) < step {
			log.Debugf("Finished sending UTXOs for pruning block %s",
				msgRequestIBDRootUTXOSetAndBlock.IBDRoot)

			return flow.outgoingRoute.Enqueue(appmessage.NewMsgDoneIBDRootUTXOSetChunks())
		}

		offset += step
		chunksSent++

		// Wait for the peer to request more chunks every `ibdBatchSize` chunks
		if chunksSent%ibdBatchSize == 0 {
			message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
			if err != nil {
				return err
			}
			_, ok := message.(*appmessage.MsgRequestNextIBDRootUTXOSetChunk)
			if !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextIBDRootUTXOSetChunk, message.Command())
			}
		}
	}
}
