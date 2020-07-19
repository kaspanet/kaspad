package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"time"
)

const minDurationToRequestSelectedTips = time.Minute

func requestSelectedTipsIfRequired(dag *blockdag.BlockDAG) {
	if isDAGTimeCurrent(dag) {
		return
	}
	requestSelectedTips()
}

func isDAGTimeCurrent(dag *blockdag.BlockDAG) bool {
	return dag.Now().Sub(dag.SelectedTipHeader().Timestamp) > minDurationToRequestSelectedTips
}

func requestSelectedTips() {
	for _, peer := range peerpkg.ReadyPeers() {
		peer.RequestSelectedTipIfRequired()
	}
}

// RequestSelectedTip waits for selected tip requests and handles them
func RequestSelectedTip(incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer, dag *blockdag.BlockDAG) error {
	for {
		peer.WaitForSelectedTipRequests()

		shouldStop, err := runSelectedTipRequest(incomingRoute, outgoingRoute, peer, dag)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
}

func runSelectedTipRequest(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) (shouldStop bool, err error) {

	defer peer.FinishRequestingSelectedTip()

	shouldStop = requestSelectedTip(outgoingRoute)
	if shouldStop {
		return true, nil
	}

	peerSelectedTipHash, shouldStop, err := receiveSelectedTip(incomingRoute)
	if err != nil {
		return true, err
	}
	if shouldStop {
		return true, nil
	}

	peer.SetSelectedTipHash(peerSelectedTipHash)
	StartIBDIfRequired(dag)
	return false, nil
}

func requestSelectedTip(outgoingRoute *router.Route) (shouldStop bool) {
	msgGetSelectedTip := wire.NewMsgGetSelectedTip()
	isOpen := outgoingRoute.Enqueue(msgGetSelectedTip)
	return !isOpen
}

func receiveSelectedTip(incomingRoute *router.Route) (selectedTipHash *daghash.Hash, shouldStop bool, err error) {
	message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, true, err
	}
	if !isOpen {
		return nil, true, nil
	}
	msgSelectedTip := message.(*wire.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, false, nil
}

// HandleGetSelectedTip handles getSelectedTip messages
func HandleGetSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		shouldStop, err := receiveGetSelectedTip(incomingRoute)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}

		selectedTipHash := dag.SelectedTipHash()
		shouldStop = sendSelectedTipHash(outgoingRoute, selectedTipHash)
		if shouldStop {
			return nil
		}
	}
}

func receiveGetSelectedTip(incomingRoute *router.Route) (shouldStop bool, err error) {
	message, isOpen := incomingRoute.Dequeue()
	if !isOpen {
		return true, nil
	}
	_, ok := message.(*wire.MsgGetSelectedTip)
	if !ok {
		panic(errors.Errorf("received unexpected message type. "+
			"expected: MsgGetSelectedTip, got: %s", message.Command()))
	}

	return false, nil
}

func sendSelectedTipHash(outgoingRoute *router.Route, selectedTipHash *daghash.Hash) (shouldStop bool) {
	msgSelectedTip := wire.NewMsgSelectedTip(selectedTipHash)
	isOpen := outgoingRoute.Enqueue(msgSelectedTip)
	return !isOpen
}
