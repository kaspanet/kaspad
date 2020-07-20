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
		err := runSelectedTipRequest(incomingRoute, outgoingRoute, peer, dag)
		if err != nil {
			return err
		}
	}
}

func runSelectedTipRequest(incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer, dag *blockdag.BlockDAG) error {

	peer.WaitForSelectedTipRequests()
	defer peer.FinishRequestingSelectedTip()

	err := requestSelectedTip(outgoingRoute)
	if err != nil {
		return err
	}

	peerSelectedTipHash, err := receiveSelectedTip(incomingRoute)
	if err != nil {
		return err
	}

	peer.SetSelectedTipHash(peerSelectedTipHash)
	StartIBDIfRequired(dag)
	return nil
}

func requestSelectedTip(outgoingRoute *router.Route) error {
	msgGetSelectedTip := wire.NewMsgGetSelectedTip()
	isOpen := outgoingRoute.Enqueue(msgGetSelectedTip)
	if !isOpen {
		return errors.WithStack(common.ErrRouteClosed)
	}
	return nil
}

func receiveSelectedTip(incomingRoute *router.Route) (selectedTipHash *daghash.Hash, err error) {
	message, isOpen, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return nil, errors.WithStack(common.ErrRouteClosed)
	}
	msgSelectedTip := message.(*wire.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, nil
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
			"expected: %s, got: %s", wire.CmdGetSelectedTip, message.Command()))
	}

	return false, nil
}

func sendSelectedTipHash(outgoingRoute *router.Route, selectedTipHash *daghash.Hash) (shouldStop bool) {
	msgSelectedTip := wire.NewMsgSelectedTip(selectedTipHash)
	isOpen := outgoingRoute.Enqueue(msgSelectedTip)
	return !isOpen
}
