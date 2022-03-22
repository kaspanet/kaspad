package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
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
			blockInfo, err := context.Domain().Consensus().GetBlockInfo(hash)
			if err != nil {
				return err
			}
			if !blockInfo.Exists || blockInfo.BlockStatus == externalapi.StatusHeaderOnly {
				return protocolerrors.Errorf(true, "block %s not found", hash)
			}
			block, err := context.Domain().Consensus().GetBlock(hash)
			if err != nil {
				return errors.Wrapf(err, "unable to fetch requested block hash %s", hash)
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
