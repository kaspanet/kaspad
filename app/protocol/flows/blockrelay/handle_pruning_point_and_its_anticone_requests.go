package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// PruningPointAndItsAnticoneRequestsContext is the interface for the context needed for the HandlePruningPointAndItsAnticoneRequests flow.
type PruningPointAndItsAnticoneRequestsContext interface {
	Domain() domain.Domain
}

// HandlePruningPointAndItsAnticoneRequests listens to appmessage.MsgRequestPruningPointAndItsAnticone messages and sends
// the pruning point and its anticone to the requesting peer.
func HandlePruningPointAndItsAnticoneRequests(context PruningPointAndItsAnticoneRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		_, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}

		log.Debugf("Got request for pruning point and its anticone from %s", peer)

		pruningPointHeaders, err := context.Domain().Consensus().PruningPointHeaders()
		if err != nil {
			return err
		}

		msgHeaders := make([]*appmessage.MsgBlockHeader, len(pruningPointHeaders))
		for i, header := range pruningPointHeaders {
			msgHeaders[i] = appmessage.DomainBlockHeaderToBlockHeader(header)
		}

		err = outgoingRoute.Enqueue(appmessage.NewMsgPruningPoints(msgHeaders))
		if err != nil {
			return err
		}

		blocks, err := context.Domain().Consensus().PruningPointAndItsAnticoneWithTrustedData()
		if err != nil {
			return err
		}

		for _, block := range blocks {
			err = outgoingRoute.Enqueue(appmessage.DomainBlockWithTrustedDataToBlockWithTrustedData(block))
			if err != nil {
				return err
			}
		}

		err = outgoingRoute.Enqueue(appmessage.NewMsgDoneBlocksWithTrustedData())
		if err != nil {
			return err
		}

		log.Debugf("Sent pruning point and its anticone to %s", peer)
	}
}
