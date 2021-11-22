package blockrelay

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/app/protocol/protocolerrors"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"runtime"
	"sync/atomic"
)

// PruningPointAndItsAnticoneRequestsContext is the interface for the context needed for the HandlePruningPointAndItsAnticoneRequests flow.
type PruningPointAndItsAnticoneRequestsContext interface {
	Domain() domain.Domain
}

var isBusy uint32

// HandlePruningPointAndItsAnticoneRequests listens to appmessage.MsgRequestPruningPointAndItsAnticone messages and sends
// the pruning point and its anticone to the requesting peer.
func HandlePruningPointAndItsAnticoneRequests(context PruningPointAndItsAnticoneRequestsContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	for {
		err := func() error {
			_, err := incomingRoute.Dequeue()
			if err != nil {
				return err
			}

			if !atomic.CompareAndSwapUint32(&isBusy, 0, 1) {
				return protocolerrors.Errorf(false, "node is busy with other pruning point anticone requests")
			}
			defer atomic.StoreUint32(&isBusy, 0)

			log.Debugf("Got request for pruning point and its anticone from %s", peer)

			pruningPointHeaders, err := context.Domain().Consensus().PruningPointHeaders()
			if err != nil {
				return err
			}

			msgPruningPointHeaders := make([]*appmessage.MsgBlockHeader, len(pruningPointHeaders))
			for i, header := range pruningPointHeaders {
				msgPruningPointHeaders[i] = appmessage.DomainBlockHeaderToBlockHeader(header)
			}

			err = outgoingRoute.Enqueue(appmessage.NewMsgPruningPoints(msgPruningPointHeaders))
			if err != nil {
				return err
			}

			pointAndItsAnticone, err := context.Domain().Consensus().PruningPointAndItsAnticone()
			if err != nil {
				return err
			}

			for _, blockHash := range pointAndItsAnticone {
				err := sendBlockWithTrustedData(context, outgoingRoute, blockHash)
				if err != nil {
					return err
				}
			}

			err = outgoingRoute.Enqueue(appmessage.NewMsgDoneBlocksWithTrustedData())
			if err != nil {
				return err
			}

			log.Debugf("Sent pruning point and its anticone to %s", peer)
			return nil
		}()
		if err != nil {
			return err
		}
	}
}

func sendBlockWithTrustedData(context PruningPointAndItsAnticoneRequestsContext, outgoingRoute *router.Route, blockHash *externalapi.DomainHash) error {
	blockWithTrustedData, err := context.Domain().Consensus().BlockWithTrustedData(blockHash)
	if err != nil {
		return err
	}

	err = outgoingRoute.Enqueue(appmessage.DomainBlockWithTrustedDataToBlockWithTrustedData(blockWithTrustedData))
	if err != nil {
		return err
	}

	runtime.GC()

	return nil
}
