package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleIBDBlockLocatorContext is the interface for the context needed for the HandleIBDBlockLocator flow.
type HandleIBDBlockLocatorContext interface {
	Domain() domain.Domain
}

// HandleIBDBlockRequests listens to appmessage.MsgIBDBlockLocator messages and sends
// the the highest known block to the requesting peer.
func HandleIBDBlockLocator(context HandleIBDBlockLocatorContext, incomingRoute *router.Route,
	outgoingRoute *router.Route) error {

	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		ibdBlockLocatorMessage := message.(*appmessage.MsgIBDBlockLocator)

		foundHighestHash := false
		for _, hash := range ibdBlockLocatorMessage.Hashes {
			blockInfo, err := context.Domain().Consensus().GetBlockInfo(hash)
			if err != nil {
				return err
			}
			if blockInfo.Exists {
				foundHighestHash = true
				ibdBlockLocatorHighestHashMessage := appmessage.NewMsgIBDBlockLocatorHighestHash(hash)
				err = outgoingRoute.Enqueue(ibdBlockLocatorHighestHashMessage)
				if err != nil {
					return err
				}
				break
			}
		}

		if !foundHighestHash {
			return protocolerrors.Errorf(true, "no known hash was found in block locator")
		}
	}
}
