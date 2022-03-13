package blockrelay

import (
	"github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// This constant must be equal at both syncer and syncee. Therefore, never (!!) change this constant unless a new p2p
// version is introduced. See `TestIBDBatchSizeLessThanRouteCapacity` as well.
const ibdBatchSize = 99

// RequestHeadersContext is the interface for the context needed for the HandleRequestHeaders flow.
type RequestHeadersContext interface {
	Domain() domain.Domain
}

type handleRequestHeadersFlow struct {
	RequestHeadersContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peer.Peer
}

// HandleRequestHeaders handles RequestHeaders messages
func HandleRequestHeaders(context RequestHeadersContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peer.Peer) error {

	flow := &handleRequestHeadersFlow{
		RequestHeadersContext: context,
		incomingRoute:         incomingRoute,
		outgoingRoute:         outgoingRoute,
		peer:                  peer,
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
			// maxBlocks MUST be >= MergeSetSizeLimit + 1
			const maxBlocks = 1 << 10
			blockHashes, _, err := flow.Domain().Consensus().GetHashesBetween(lowHash, highHash, maxBlocks)
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
