package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"time"
)

const (
	minDurationToRequestSelectedTips = time.Minute
)

func requestSelectedTipsIfRequired(dag *blockdag.BlockDAG) error {
	if hasRecentlyReceivedBlock(dag) {
		return nil
	}
	return requestSelectedTips(dag)
}

func hasRecentlyReceivedBlock(dag *blockdag.BlockDAG) bool {
	return dag.Now().Sub(dag.SelectedTipHeader().Timestamp) > minDurationToRequestSelectedTips
}

func requestSelectedTips(dag *blockdag.BlockDAG) error {
	for _, peer := range peerpkg.ReadyPeers() {
		peer.RequestSelectedTipIfRequired()
	}
	return nil
}

func RequestSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route, peer *peerpkg.Peer) error {
	for {
		peer.WaitForSelectedTipRequests()

		// We run the flow inside a func so that the defer is called at its end
		err := func() error {
			defer peer.FinishRequestingSelectedTip()

			shouldContinue, err := requestSelectedTip(outgoingRoute)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			peerSelectedTipHash, shouldContinue, err := receiveSelectedTip(incomingRoute)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
			return peer.SetSelectedTipHash(peerSelectedTipHash)
		}()
		if err != nil {
			return err
		}
	}
}

func requestSelectedTip(outgoingRoute *router.Route) (shouldContinue bool, err error) {
	msgGetSelectedTip := wire.NewMsgGetSelectedTip()
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgGetSelectedTip, common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	return isOpen, nil
}

func receiveSelectedTip(incomingRoute *router.Route) (selectedTipHash *daghash.Hash, shouldContinue bool, err error) {
	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return nil, false, nil
	}
	msgSelectedTip := message.(*wire.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, true, nil
}

func HandleGetSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {
	return nil
}
