package blockrelay

import (
	"github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// RequestPastDiffContext is the interface for the context needed for the HandleRequestHeaders flow.
type RequestPastDiffContext interface {
	Domain() domain.Domain
}

type handleRequestPastDiffFlow struct {
	RequestPastDiffContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peer.Peer
}

// HandleRequestPastDiff handles RequestPastDiff messages
func HandleRequestPastDiff(context RequestPastDiffContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peer.Peer) error {

	flow := &handleRequestPastDiffFlow{
		RequestPastDiffContext: context,
		incomingRoute:          incomingRoute,
		outgoingRoute:          outgoingRoute,
		peer:                   peer,
	}
	return flow.start()
}

func (flow *handleRequestPastDiffFlow) start() error {
	for {
		hasHash, requestedHash, err := receiveRequestPastDiff(flow.incomingRoute)
		if err != nil {
			return err
		}
		log.Debugf("Received requestPastDiff with hasHash: %s, requestedHash: %s", hasHash, requestedHash)

		// TODO: implement logic
		for !hasHash.Equal(requestedHash) {
			log.Debugf("Getting block headers between %s and %s to %s", hasHash, requestedHash, flow.peer)

			// GetHashesBetween is a relatively heavy operation so we limit it
			// in order to avoid locking the consensus for too long
			// maxBlocks MUST be >= MergeSetSizeLimit + 1
			const maxBlocks = 1 << 10
			// TODO: implement past diff API
			blockHashes, _, err := flow.Domain().Consensus().GetHashesBetween(hasHash, requestedHash, maxBlocks)
			if err != nil {
				return err
			}
			log.Debugf("Got %d header hashes above hasHash %s", len(blockHashes), hasHash)

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

			// The next hasHash is the last element in blockHashes
			hasHash = blockHashes[len(blockHashes)-1]
		}

		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgDoneHeaders())
		if err != nil {
			return err
		}
	}
}

func receiveRequestPastDiff(incomingRoute *router.Route) (hasHash *externalapi.DomainHash,
	requestedHash *externalapi.DomainHash, err error) {

	message, err := incomingRoute.Dequeue()
	if err != nil {
		return nil, nil, err
	}
	msgRequestPastDiff := message.(*appmessage.MsgRequestPastDiff)

	return msgRequestPastDiff.HasHash, msgRequestPastDiff.RequestedHash, nil
}
