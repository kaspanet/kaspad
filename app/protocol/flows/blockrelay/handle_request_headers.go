package blockrelay

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const ibdBatchSize = router.DefaultMaxMessages

// RequestIBDBlocksContext is the interface for the context needed for the HandleRequestHeaders flow.
type RequestIBDBlocksContext interface {
	Domain() domain.Domain
}

type handleRequestBlocksFlow struct {
	RequestIBDBlocksContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestHeaders handles RequestHeaders messages
func HandleRequestHeaders(context RequestIBDBlocksContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRequestBlocksFlow{
		RequestIBDBlocksContext: context,
		incomingRoute:           incomingRoute,
		outgoingRoute:           outgoingRoute,
	}
	return flow.start()
}

func (flow *handleRequestBlocksFlow) start() error {
	for {
		lowHash, highHash, err := receiveRequestHeaders(flow.incomingRoute)
		if err != nil {
			return err
		}

		msgHeaders, err := flow.buildMsgBlockHeaders(lowHash, highHash)
		if err != nil {
			return err
		}

		for offset := 0; offset < len(msgHeaders); offset += ibdBatchSize {
			end := offset + ibdBatchSize
			if end > len(msgHeaders) {
				end = len(msgHeaders)
			}

			blocksToSend := msgHeaders[offset:end]
			err = flow.sendHeaders(blocksToSend)
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

			if _, ok := message.(*appmessage.MsgRequestNextHeaders); !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextHeaders, message.Command())
			}
		}
		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
		if err != nil {
			return err
		}
	}
}

func receiveRequestHeaders(incomingRoute *router.Route) (lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash, err error) {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgRequestIBDBlocks := message.(*appmessage.MsgRequestHeaders)

	return msgRequestIBDBlocks.LowHash, msgRequestIBDBlocks.HighHash, nil
}

func (flow *handleRequestBlocksFlow) buildMsgBlockHeaders(lowHash *externalapi.DomainHash,
	highHash *externalapi.DomainHash) ([]*appmessage.MsgBlockHeader, error) {

	blockHashes, err := flow.Domain().Consensus().GetHashesBetween(lowHash, highHash)
	if err != nil {
		return nil, err
	}

	msgBlockHeaders := make([]*appmessage.MsgBlockHeader, len(blockHashes))
	for i, blockHash := range blockHashes {
		header, err := flow.Domain().Consensus().GetBlockHeader(blockHash)
		if err != nil {
			return nil, err
		}
		msgBlockHeaders[i] = appmessage.DomainBlockHeaderToBlockHeader(header)
	}

	return msgBlockHeaders, nil
}

func (flow *handleRequestBlocksFlow) sendHeaders(headers []*appmessage.MsgBlockHeader) error {
	for _, msgBlockHeader := range headers {
		err := flow.outgoingRoute.Enqueue(msgBlockHeader)
		if err != nil {
			return err
		}
	}
	return nil
}
