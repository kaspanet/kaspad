package blockrelay

import (
	"github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

const ibdBatchSize = router.DefaultMaxMessages

// RequestIBDBlocksContext is the interface for the context needed for the HandleRequestIBDBlocks flow.
type RequestIBDBlocksContext interface {
	Domain() domain.Domain
}

type handleRequestIBDBlocksFlow struct {
	RequestIBDBlocksContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peer.Peer
}

// HandleRequestIBDBlocks handles RequestIBDBlocks messages
func HandleRequestIBDBlocks(context RequestIBDBlocksContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peer.Peer) error {

	flow := &handleRequestIBDBlocksFlow{
		RequestIBDBlocksContext: context,
		incomingRoute:           incomingRoute,
		outgoingRoute:           outgoingRoute,
		peer:                    peer,
	}
	return flow.start()
}

func (flow *handleRequestIBDBlocksFlow) start() error {
	for {
		lowHash, highHash, err := receiveRequestIBDBlocks(flow.incomingRoute)
		if err != nil {
			return err
		}
		log.Debugf("Received requestIBDBlocks with lowHash: %s, highHash: %s", lowHash, highHash)

		for !lowHash.Equal(highHash) {
			log.Debugf("Getting block headers between %s and %s to %s", lowHash, highHash, flow.peer)

			// GetHashesBetween is a relatively heavy operation so we limit it
			// in order to avoid locking the consensus for too long
			const maxBlueScoreDifference = 1 << 10
			blockHashes, _, err := flow.Domain().Consensus().GetHashesBetween(lowHash, highHash, maxBlueScoreDifference)
			if err != nil {
				return err
			}
			log.Debugf("Got %d hashes above lowHash %s", len(blockHashes), lowHash)

			blocks := make([]*appmessage.MsgBlock, len(blockHashes))
			for i, blockHash := range blockHashes {
				blockHeader, err := flow.Domain().Consensus().GetBlock(blockHash)
				if err != nil {
					return err
				}
				blocks[i] = appmessage.DomainBlockToMsgBlock(blockHeader)
			}

			blockHeadersMessage := appmessage.NewBlockHeadersMessage(blocks)
			err = flow.outgoingRoute.Enqueue(blockHeadersMessage)
			if err != nil {
				return err
			}

			message, err := flow.incomingRoute.Dequeue()
			if err != nil {
				return err
			}
			if _, ok := message.(*appmessage.MsgRequestNextIBDBlocks); !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextIBDBlocks, message.Command())
			}

			// The next lowHash is the last element in blockHashes
			lowHash = blockHashes[len(blockHashes)-1]
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
