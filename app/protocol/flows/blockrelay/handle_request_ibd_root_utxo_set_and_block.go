package blockrelay

import (
	"errors"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
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
		message, err := flow.incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		msgRequestIBDRootUTXOSetAndBlock, ok := message.(*appmessage.MsgRequestIBDRootUTXOSetAndBlock)
		if !ok {
			return protocolerrors.Errorf(true, "received unexpected message type. "+
				"expected: %s, got: %s", appmessage.CmdRequestIBDRootUTXOSetAndBlock, message.Command())
		}

		log.Debugf("Got request for IBDRoot UTXOSet and Block")

		serializedUTXOSet, err := flow.Domain().Consensus().GetPruningPointUTXOSet(msgRequestIBDRootUTXOSetAndBlock.IBDRoot)
		if err != nil {
			if errors.Is(err, ruleerrors.ErrWrongPruningPointHash) {
				err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDRootNotFound())
				if err != nil {
					return err
				}

				continue
			}
		}
		log.Debugf("Retrieved utxo set for pruning block %s", msgRequestIBDRootUTXOSetAndBlock.IBDRoot)

		block, err := flow.Domain().Consensus().GetBlock(msgRequestIBDRootUTXOSetAndBlock.IBDRoot)
		if err != nil {
			return err
		}
		log.Debugf("Retrieved pruning block %s", msgRequestIBDRootUTXOSetAndBlock.IBDRoot)

		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(block)))
		if err != nil {
			return err
		}

		// Send the UTXO set in `step`-sized chunks
		const step = 1024 * 1024 * 1024 // 1MB
		offset := 0
		chunksSent := 0
		for offset < len(serializedUTXOSet) {
			var chunk []byte
			if offset+step < len(serializedUTXOSet) {
				chunk = serializedUTXOSet[offset : offset+step]
			} else {
				chunk = serializedUTXOSet[offset:]
			}

			err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDRootUTXOSetChunk(chunk))
			if err != nil {
				return err
			}

			offset += step
			chunksSent++

			// Wait for the peer to request more chunks every `ibdBatchSize` chunks
			if chunksSent%ibdBatchSize == 0 {
				message, err := flow.outgoingRoute.DequeueWithTimeout(common.DefaultTimeout)
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

		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgDoneIBDRootUTXOSetChunks())
		if err != nil {
			return err
		}
	}
}
