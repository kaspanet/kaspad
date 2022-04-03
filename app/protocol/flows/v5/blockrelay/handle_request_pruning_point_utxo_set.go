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

// HandleRequestPruningPointUTXOSetContext is the interface for the context needed for the HandleRequestPruningPointUTXOSet flow.
type HandleRequestPruningPointUTXOSetContext interface {
	Domain() domain.Domain
}

type handleRequestPruningPointUTXOSetFlow struct {
	HandleRequestPruningPointUTXOSetContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestPruningPointUTXOSet listens to appmessage.MsgRequestPruningPointUTXOSet messages and sends
// the pruning point UTXO set and block body.
func HandleRequestPruningPointUTXOSet(context HandleRequestPruningPointUTXOSetContext, incomingRoute,
	outgoingRoute *router.Route) error {

	flow := &handleRequestPruningPointUTXOSetFlow{
		HandleRequestPruningPointUTXOSetContext: context,
		incomingRoute:                           incomingRoute,
		outgoingRoute:                           outgoingRoute,
	}

	return flow.start()
}

func (flow *handleRequestPruningPointUTXOSetFlow) start() error {
	for {
		msgRequestPruningPointUTXOSet, err := flow.waitForRequestPruningPointUTXOSetMessages()
		if err != nil {
			return err
		}

		err = flow.handleRequestPruningPointUTXOSetMessage(msgRequestPruningPointUTXOSet)
		if err != nil {
			return err
		}
	}
}

func (flow *handleRequestPruningPointUTXOSetFlow) handleRequestPruningPointUTXOSetMessage(
	msgRequestPruningPointUTXOSet *appmessage.MsgRequestPruningPointUTXOSet) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "handleRequestPruningPointUTXOSetFlow")
	defer onEnd()

	log.Tracef("Got request for pruning point UTXO set")

	return flow.sendPruningPointUTXOSet(msgRequestPruningPointUTXOSet)
}

func (flow *handleRequestPruningPointUTXOSetFlow) waitForRequestPruningPointUTXOSetMessages() (
	*appmessage.MsgRequestPruningPointUTXOSet, error) {

	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return nil, err
	}
	msgRequestPruningPointUTXOSet, ok := message.(*appmessage.MsgRequestPruningPointUTXOSet)
	if !ok {
		// TODO: Change to shouldBan: true once we fix the bug of getting redundant messages
		return nil, protocolerrors.Errorf(false, "received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdRequestPruningPointUTXOSet, message.Command())
	}
	return msgRequestPruningPointUTXOSet, nil
}

func (flow *handleRequestPruningPointUTXOSetFlow) sendPruningPointUTXOSet(
	msgRequestPruningPointUTXOSet *appmessage.MsgRequestPruningPointUTXOSet) error {

	// Send the UTXO set in `step`-sized chunks
	const step = 1000
	var fromOutpoint *externalapi.DomainOutpoint
	chunksSent := 0
	for {
		pruningPointUTXOs, err := flow.Domain().Consensus().GetPruningPointUTXOs(
			msgRequestPruningPointUTXOSet.PruningPointHash, fromOutpoint, step)
		if err != nil {
			if errors.Is(err, ruleerrors.ErrWrongPruningPointHash) {
				return flow.outgoingRoute.Enqueue(appmessage.NewMsgUnexpectedPruningPoint())
			}
		}

		log.Tracef("Retrieved %d UTXOs for pruning block %s",
			len(pruningPointUTXOs), msgRequestPruningPointUTXOSet.PruningPointHash)

		outpointAndUTXOEntryPairs :=
			appmessage.DomainOutpointAndUTXOEntryPairsToOutpointAndUTXOEntryPairs(pruningPointUTXOs)
		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgPruningPointUTXOSetChunk(outpointAndUTXOEntryPairs))
		if err != nil {
			return err
		}

		finished := len(pruningPointUTXOs) < step
		if finished && chunksSent%ibdBatchSize != 0 {
			log.Tracef("Finished sending UTXOs for pruning block %s",
				msgRequestPruningPointUTXOSet.PruningPointHash)

			return flow.outgoingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
		}

		if len(pruningPointUTXOs) > 0 {
			fromOutpoint = pruningPointUTXOs[len(pruningPointUTXOs)-1].Outpoint
		}
		chunksSent++

		// Wait for the peer to request more chunks every `ibdBatchSize` chunks
		if chunksSent%ibdBatchSize == 0 {
			message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
			if err != nil {
				return err
			}
			_, ok := message.(*appmessage.MsgRequestNextPruningPointUTXOSetChunk)
			if !ok {
				// TODO: Change to shouldBan: true once we fix the bug of getting redundant messages
				return protocolerrors.Errorf(false, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextPruningPointUTXOSetChunk, message.Command())
			}

			if finished {
				log.Tracef("Finished sending UTXOs for pruning block %s",
					msgRequestPruningPointUTXOSet.PruningPointHash)

				return flow.outgoingRoute.Enqueue(appmessage.NewMsgDonePruningPointUTXOSetChunks())
			}
		}
	}
}
