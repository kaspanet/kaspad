package ibd

import (
	"errors"
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

// RequestSelectedTip waits for selected tip requests and handles them
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
	message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, false, err
	}
	if !isOpen {
		return nil, false, nil
	}
	msgSelectedTip := message.(*wire.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, true, nil
}

// HandleGetSelectedTip handles getSelectedTip messages
func HandleGetSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		shouldContinue, err := receiveGetSelectedTip(incomingRoute)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}

		selectedTipHash := dag.SelectedTipHash()
		shouldContinue, err = sendSelectedTipHash(outgoingRoute, selectedTipHash)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}
	}
}

func receiveGetSelectedTip(incomingRoute *router.Route) (shouldContinue bool, err error) {
	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return false, nil
	}
	_, ok := message.(*wire.MsgGetSelectedTip)
	if !ok {
		panic(errors.New("received unexpected message type"))
	}

	return true, nil
}

func sendSelectedTipHash(outgoingRoute *router.Route, selectedTipHash *daghash.Hash) (shouldContinue bool, err error) {
	msgSelectedTip := wire.NewMsgSelectedTip(selectedTipHash)
	isOpen, err := outgoingRoute.EnqueueWithTimeout(msgSelectedTip, common.DefaultTimeout)
	if err != nil {
		return false, err
	}
	return isOpen, nil
}
