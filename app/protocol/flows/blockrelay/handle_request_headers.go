package blockrelay

import (
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
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

		batchBlockHeaders := make([]*appmessage.MsgBlockHeader, 0, ibdBatchSize)
		for !lowHash.Equal(highHash) {
			// GetHashesBetween is a relatively heavy operation so we limit it
			// in order to avoid locking the consensus for too long
			const maxBlueScoreDifference = 1 << 10
			blockHashes, err := flow.Domain().Consensus().GetHashesBetween(lowHash, highHash, maxBlueScoreDifference)
			if err != nil {
				return err
			}

			offset := 0
			for offset < len(blockHashes) {
				for len(batchBlockHeaders) < ibdBatchSize {
					hashAtOffset := blockHashes[offset]
					blockHeader, err := flow.Domain().Consensus().GetBlockHeader(hashAtOffset)
					if err != nil {
						return err
					}
					blockHeaderMessage := appmessage.DomainBlockHeaderToBlockHeader(blockHeader)
					batchBlockHeaders = append(batchBlockHeaders, blockHeaderMessage)

					offset++
					if offset == len(blockHashes) {
						break
					}
				}

				if len(batchBlockHeaders) < ibdBatchSize {
					break
				}

				err = flow.sendHeaders(batchBlockHeaders)
				if err != nil {
					return nil
				}
				batchBlockHeaders = make([]*appmessage.MsgBlockHeader, 0, ibdBatchSize)

				message, err := flow.incomingRoute.Dequeue()
				if err != nil {
					return err
				}
				if _, ok := message.(*appmessage.MsgRequestNextHeaders); !ok {
					return protocolerrors.Errorf(true, "received unexpected message type. "+
						"expected: %s, got: %s", appmessage.CmdRequestNextHeaders, message.Command())
				}
			}

			// The next lowHash is the last element in blockHashes
			lowHash = blockHashes[len(blockHashes)-1]
		}

		if len(batchBlockHeaders) > 0 {
			err = flow.sendHeaders(batchBlockHeaders)
			if err != nil {
				return nil
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

func (flow *handleRequestBlocksFlow) sendHeaders(headers []*appmessage.MsgBlockHeader) error {
	for _, msgBlockHeader := range headers {
		err := flow.outgoingRoute.Enqueue(msgBlockHeader)
		if err != nil {
			return err
		}
	}
	return nil
}
