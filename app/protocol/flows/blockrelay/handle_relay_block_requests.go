package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// RelayBlockRequestsContext is the interface for the context needed for the HandleRelayBlockRequests flow.
type RelayBlockRequestsContext interface {
	Domain() domain.Domain
}

// HandleRelayBlockRequests listens to appmessage.MsgRequestRelayBlocks messages and sends
// their corresponding blocks to the requesting peer.
func HandleRelayBlockRequests(context RelayBlockRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		getRelayBlocksMessage := message.(*appmessage.MsgRequestRelayBlocks)
		for _, hash := range getRelayBlocksMessage.Hashes {
			// Fetch the block from the database.
			block, err := context.Domain().GetBlock(hash)
			if blockdag.IsNotInDAGErr(err) {
				return protocolerrors.Errorf(true, "block %s not found", hash)
			} else if err != nil {
				return errors.Wrapf(err, "unable to fetch requested block hash %s", hash)
			}

			// TODO (Partial nodes): Convert block to partial block if needed

			DomainBlock

			err = outgoingRoute.Enqueue(msgBlock)
			if err != nil {
				return err
			}
		}
	}
}
