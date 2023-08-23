package blockrelay

import (
	"github.com/c4ei/kaspad/app/appmessage"
	"github.com/c4ei/kaspad/app/protocol/protocolerrors"
	"github.com/c4ei/kaspad/domain"
	"github.com/c4ei/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleIBDBlockRequestsContext is the interface for the context needed for the HandleIBDBlockRequests flow.
type HandleIBDBlockRequestsContext interface {
	Domain() domain.Domain
}

// HandleIBDBlockRequests listens to appmessage.MsgRequestRelayBlocks messages and sends
// their corresponding blocks to the requesting peer.
func HandleIBDBlockRequests(context HandleIBDBlockRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route) error {

	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		msgRequestIBDBlocks := message.(*appmessage.MsgRequestIBDBlocks)
		log.Debugf("Got request for %d ibd blocks", len(msgRequestIBDBlocks.Hashes))
		for i, hash := range msgRequestIBDBlocks.Hashes {
			// Fetch the block from the database.
			block, found, err := context.Domain().Consensus().GetBlock(hash)
			if err != nil {
				return errors.Wrapf(err, "unable to fetch requested block hash %s", hash)
			}

			if !found {
				return protocolerrors.Errorf(false, "IBD block %s not found", hash)
			}

			// TODO (Partial nodes): Convert block to partial block if needed

			blockMessage := appmessage.DomainBlockToMsgBlock(block)
			ibdBlockMessage := appmessage.NewMsgIBDBlock(blockMessage)
			err = outgoingRoute.Enqueue(ibdBlockMessage)
			if err != nil {
				return err
			}
			log.Debugf("sent %d out of %d", i+1, len(msgRequestIBDBlocks.Hashes))
		}
	}
}
