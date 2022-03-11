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
		log.Debugf("Getting past(%s) setminus past(%s) to %s", requestedHash, hasHash, flow.peer)

		// GetPastDiff is a relatively heavy operation so we limit it
		// in order to avoid locking the consensus for too long
		// maxBlocks MUST be >= MergeSetSizeLimit + 1
		const maxBlocks = 1 << 10
		blockHashes, err := flow.Domain().Consensus().GetPastDiff(hasHash, requestedHash, maxBlocks)
		if err != nil {
			return protocolerrors.Wrap(true, err, "Expected hashes in anticone one of the other")
		}
		log.Debugf("Got %d header hashes in past(%s) setminus past(%s)", len(blockHashes), requestedHash, hasHash)

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
