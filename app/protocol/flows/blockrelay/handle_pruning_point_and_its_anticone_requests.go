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

// HandleBlockBlueWorkRequests listens to appmessage.MsgRequestPruningPointAndItsAnticone messages and sends
// the pruning point and its anticone to the requesting peer.
func HandlePruningPointAndItsAnticoneRequests(context PruningPointAndItsAnticoneRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		_, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		log.Debugf("Got request for pruning point and its anticone from %s", peer)
		blocks, err := context.Domain().Consensus().PruningPointAndItsAnticoneWithMetaData()
		if err != nil {
			return err
		}

		for _, block := range blocks {
			err = outgoingRoute.Enqueue(appmessage.DomainBlockWithMetaDataToBlockWithMetaData(block))
			if err != nil {
				return err
			}
		}

		log.Debugf("Sent pruning point and its anticone to %s", peer)
	}
}
