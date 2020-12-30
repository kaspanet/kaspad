package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleIBDBlockLocatorContext is the interface for the context needed for the HandleIBDBlockLocator flow.
type HandleIBDBlockLocatorContext interface {
	Domain() domain.Domain
}

// HandleIBDBlockRequests listens to appmessage.MsgIBDBlockLocator messages and sends
// the highest known block that's in the selected parent chain of `targetHash` to the
// requesting peer.
func HandleIBDBlockLocator(context HandleIBDBlockLocatorContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peer.Peer) error {

	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		ibdBlockLocatorMessage := message.(*appmessage.MsgIBDBlockLocator)

		targetHash := ibdBlockLocatorMessage.TargetHash
		log.Debugf("Received IBDBlockLocator from %s with targetHash %s", peer, targetHash)

		blockInfo, err := context.Domain().Consensus().GetBlockInfo(targetHash)
		if err != nil {
			return err
		}
		if !blockInfo.Exists {
			return protocolerrors.Errorf(true, "received IBDBlockLocator "+
				"with an unknown targetHash %s", targetHash)
		}

		foundHighestHashInTheSelectedParentChainOfTargetHash := false
		for _, blockLocatorHash := range ibdBlockLocatorMessage.BlockLocatorHashes {
			blockInfo, err := context.Domain().Consensus().GetBlockInfo(blockLocatorHash)
			if err != nil {
				return err
			}
			if !blockInfo.Exists {
				continue
			}

			isBlockLocatorHashInSelectedParentChainOfHighHash, err :=
				context.Domain().Consensus().IsInSelectedParentChainOf(blockLocatorHash, targetHash)
			if err != nil {
				return err
			}
			if !isBlockLocatorHashInSelectedParentChainOfHighHash {
				continue
			}

			foundHighestHashInTheSelectedParentChainOfTargetHash = true
			log.Debugf("Found a known hash %s amongst peer %s's "+
				"blockLocator that's in the selected parent chain of targetHash %s", blockLocatorHash, peer, targetHash)

			ibdBlockLocatorHighestHashMessage := appmessage.NewMsgIBDBlockLocatorHighestHash(blockLocatorHash)
			err = outgoingRoute.Enqueue(ibdBlockLocatorHighestHashMessage)
			if err != nil {
				return err
			}
			break
		}

		if !foundHighestHashInTheSelectedParentChainOfTargetHash {
			return protocolerrors.Errorf(true, "no hash was found in the blockLocator "+
				"that was in the selected parent chain of targetHash %s", targetHash)
		}
	}
}
