package blockrelay

import (
	"errors"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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

// HandleRequestPruningPointUTXOSetAndBlock listens to appmessage.MsgRequestPruningPointUTXOSet messages and sends
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
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSet) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "handleRequestPruningPointUTXOSetAndBlockFlow")
	defer onEnd()

	log.Debugf("Got request for pruning point UTXO set")

	return flow.sendPruningPointUTXOSet(msgRequestPruningPointUTXOSetAndBlock)
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) waitForRequestPruningPointUTXOSetAndBlockMessages() (
	*appmessage.MsgRequestPruningPointUTXOSet, error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}
	msgRequestPruningPointUTXOSetAndBlock, ok := message.(*appmessage.MsgRequestPruningPointUTXOSet)
	if !ok {
		return nil, protocolerrors.Errorf(true, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdRequestPruningPointUTXOSet, message.Command())
	}
	return msgRequestPruningPointUTXOSetAndBlock, nil
}

func (flow *handleRequestPruningPointUTXOSetAndBlockFlow) sendPruningPointUTXOSet(
	msgRequestPruningPointUTXOSetAndBlock *appmessage.MsgRequestPruningPointUTXOSet) error {

	// Send the UTXO set in `step`-sized chunks
	const step = 1000
	var fromOutpoint *externalapi.DomainOutpoint
	chunksSent := 0
	for {
		pruningPointUTXOs, err := flow.Domain().Consensus().GetPruningPointUTXOs(
			msgRequestPruningPointUTXOSetAndBlock.PruningPointHash, fromOutpoint, step)
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

		fromOutpoint = pruningPointUTXOs[len(pruningPointUTXOs)-1].Outpoint
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
