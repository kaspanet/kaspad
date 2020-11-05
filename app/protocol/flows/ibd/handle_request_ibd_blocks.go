package ibd

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const ibdBatchSize = router.DefaultMaxMessages

// RequestIBDBlocksContext is the interface for the context needed for the HandleRequestIBDBlocks flow.
type RequestIBDBlocksContext interface {
	Domain() domain.Domain
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

		for offset := 0; offset < len(msgIBDBlocks); offset += ibdBatchSize {
			end := offset + ibdBatchSize
			if end > len(msgIBDBlocks) {
				end = len(msgIBDBlocks)
			}

			blocksToSend := msgIBDBlocks[offset:end]
			err = flow.sendMsgIBDBlocks(blocksToSend)
			if err != nil {
				return nil
			}

			// Exit the loop and don't wait for the GetNextIBDBlocks message if the last batch was
			// less than ibdBatchSize.
			if len(blocksToSend) < ibdBatchSize {
				break
			}

			message, err := flow.incomingRoute.Dequeue()
			if err != nil {
				return err
			}

			if _, ok := message.(*appmessage.MsgRequestNextIBDBlocks); !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextIBDBlocks, message.Command())
			}
		}
		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgDoneIBDBlocks())
		if err != nil {
			return err
		}
	}
}

func receiveRequestIBDBlocks(incomingRoute *router.Route) (lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash, err error) {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgRequestIBDBlocks := message.(*appmessage.MsgRequestIBDBlocks)

	return msgRequestIBDBlocks.LowHash, msgRequestIBDBlocks.HighHash, nil
}

func (flow *handleRequestBlocksFlow) buildMsgIBDBlocks(lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash) ([]*appmessage.MsgIBDBlock, error) {

	blockHashes, err := flow.Domain().Consensus().GetHashesBetween(lowHash, highHash)
	if err != nil {
		return nil, err
	}
	const maxHashesInMsgIBDBlocks = appmessage.MaxInvPerMsg
	if len(blockHashes) > maxHashesInMsgIBDBlocks {
		blockHashes = blockHashes[:maxHashesInMsgIBDBlocks]
	}

	msgIBDBlocks := make([]*appmessage.MsgIBDBlock, len(blockHashes))
	for i, blockHash := range blockHashes {
		block, err := flow.Domain().Consensus().GetBlock(blockHash)
		if err != nil {
			return nil, err
		}
		msgIBDBlocks[i] = appmessage.NewMsgIBDBlock(appmessage.DomainBlockToMsgBlock(block))
	}

	return msgIBDBlocks, nil
}

func (flow *handleRequestBlocksFlow) sendMsgIBDBlocks(msgIBDBlocks []*appmessage.MsgIBDBlock) error {
	for _, msgIBDBlock := range msgIBDBlocks {
		err := flow.outgoingRoute.Enqueue(msgIBDBlock)
		if err != nil {
			return err
		}
	}
	return nil
}
