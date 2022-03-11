package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"sort"
)

// RequestPastDiffContext is the interface for the context needed for the HandleRequestHeaders flow.
type RequestPastDiffContext interface {
	Domain() domain.Domain
	Config() *config.Config
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

		// GetPastDiff is expected to be called by the syncee for getting the anticone of the header selected tip
		// intersected by past of relayed block, and is thus expected to be bounded by mergeset limit since
		// we relay blocks only if they enter virtual's mergeset. We add 2 for a small margin error.
		blockHashes, err := flow.Domain().Consensus().GetPastDiff(hasHash, requestedHash,
			flow.Config().ActiveNetParams.MergeSetSizeLimit+2)
		if err != nil {
			return protocolerrors.Wrap(true, err, "Failed querying anticone")
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

		// We sort the headers in bottom-up topological order before sending
		sort.Slice(blockHeaders, func(i, j int) bool {
			return blockHeaders[i].BlueWork.Cmp(blockHeaders[j].BlueWork) < 0
		})
		if err != nil {
			return err
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
