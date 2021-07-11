package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// BlockBlueWorkRequestsContext is the interface for the context needed for the HandleBlockBlueWorkRequests flow.
type BlockBlueWorkRequestsContext interface {
	Domain() domain.Domain
}

// HandleBlockBlueWorkRequests listens to appmessage.MsgRequestBlockBlueWork messages and sends
// their corresponding blue work to the requesting peer.
func HandleBlockBlueWorkRequests(context BlockBlueWorkRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		message, err := incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		msgRequestBlockBlueWork := message.(*appmessage.MsgRequestBlockBlueWork)
		log.Debugf("Got request for block %s blue work from %s", msgRequestBlockBlueWork.Hash, peer)
		blockInfo, err := context.Domain().Consensus().GetBlockInfo(msgRequestBlockBlueWork.Hash)
		if err != nil {
			return err
		}
		if !blockInfo.Exists {
			return protocolerrors.Errorf(true, "block %s not found", msgRequestBlockBlueWork.Hash)
		}

		err = outgoingRoute.Enqueue(appmessage.NewBlockBlueWork(blockInfo.BlueWork))
		if err != nil {
			return err
		}
		log.Debugf("Sent blue work for block %s to %s", msgRequestBlockBlueWork.Hash, peer)
	}
}
