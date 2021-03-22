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

// RequestIBDBlocksContext is the interface for the context needed for the HandleRequestHeaders flow.
type RequestIBDBlocksContext interface {
	Domain() domain.Domain
}

type handleRequestHeadersFlow struct {
	RequestIBDBlocksContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peer.Peer
}

// HandleRequestHeaders handles RequestHeaders messages
func HandleRequestHeaders(context RequestIBDBlocksContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peer.Peer) error {

	flow := &handleRequestHeadersFlow{
		RequestIBDBlocksContext: context,
		incomingRoute:           incomingRoute,
		outgoingRoute:           outgoingRoute,
		peer:                    peer,
	}
	return flow.start()
}

func (flow *handleRequestHeadersFlow) start() error {
	for {
		lowHash, highHash, err := receiveRequestHeaders(flow.incomingRoute)
		if err != nil {
			return err
		}
		log.Debugf("Recieved requestHeaders with lowHash: %s, highHash: %s", lowHash, highHash)

		for !lowHash.Equal(highHash) {
			log.Debugf("Getting block headers between %s and %s to %s", lowHash, highHash, flow.peer)

			// GetHashesBetween is a relatively heavy operation so we limit it
			// in order to avoid locking the consensus for too long
			const maxBlueScoreDifference = 1 << 10
			blockHashes, _, err := flow.Domain().Consensus().GetHashesBetween(lowHash, highHash, maxBlueScoreDifference)
			if err != nil {
				return err
			}
			log.Debugf("Got %d header hashes above lowHash %s", len(blockHashes), lowHash)

			blockHeaders := make([]*appmessage.MsgBlockHeader, len(blockHashes))
			for i, blockHash := range blockHashes {
				blockHeader, err := flow.Domain().Consensus().GetBlockHeader(blockHash)
				if err != nil {
					return err
				}
				blockHeaders[i] = appmessage.DomainBlockHeaderToBlockHeader(blockHeader)
			}

			blockHeadersMessage := appmessage.NewBlockHeadersMessage(blockHeaders)
			err = flow.outgoingRoute.Enqueue(blockHeadersMessage)
			if err != nil {
				return err
			}

			message, err := flow.incomingRoute.Dequeue()
			if err != nil {
				return err
			}
			if _, ok := message.(*appmessage.MsgRequestNextHeaders); !ok {
				return protocolerrors.Errorf(true, "received unexpected message type. "+
					"expected: %s, got: %s", appmessage.CmdRequestNextHeaders, message.Command())
			}

			// The next lowHash is the last element in blockHashes
			lowHash = blockHashes[len(blockHashes)-1]
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
