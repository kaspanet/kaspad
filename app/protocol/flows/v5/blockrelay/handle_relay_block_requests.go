package blockrelay

import (
	"github.com/c4ei/yunseokyeol/app/appmessage"
	peerpkg "github.com/c4ei/yunseokyeol/app/protocol/peer"
	"github.com/c4ei/yunseokyeol/app/protocol/protocolerrors"
	"github.com/c4ei/yunseokyeol/domain"
	"github.com/c4ei/yunseokyeol/infrastructure/network/netadapter/router"
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
		log.Debugf("Got request for relay blocks with hashes %s", getRelayBlocksMessage.Hashes)
		for _, hash := range getRelayBlocksMessage.Hashes {
			// Fetch the block from the database.
			block, found, err := context.Domain().Consensus().GetBlock(hash)
			if err != nil {
				return errors.Wrapf(err, "unable to fetch requested block hash %s", hash)
			}

			if !found {
				return protocolerrors.Errorf(false, "Relay block %s not found", hash)
			}

			// TODO (Partial nodes): Convert block to partial block if needed

			err = outgoingRoute.Enqueue(appmessage.DomainBlockToMsgBlock(block))
			if err != nil {
				return err
			}
			log.Debugf("Relayed block with hash %s", hash)
		}
	}
}
