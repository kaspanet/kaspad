package ibd

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/protocol/peer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

type RequestSelectedTipContext interface {
	DAG() *blockdag.BlockDAG
	StartIBDIfRequired()
}

// RequestSelectedTip waits for selected tip requests and handles them
func RequestSelectedTip(context RequestSelectedTipContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {
	for {
		err := runSelectedTipRequest(context, incomingRoute, outgoingRoute, peer)
		if err != nil {
			return err
		}
	}
}

func runSelectedTipRequest(context RequestSelectedTipContext, incomingRoute *router.Route, outgoingRoute *router.Route,
	peer *peerpkg.Peer) error {

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
	context.StartIBDIfRequired()
	return nil
}

func requestSelectedTip(outgoingRoute *router.Route) error {
	msgGetSelectedTip := wire.NewMsgGetSelectedTip()
	return outgoingRoute.Enqueue(msgGetSelectedTip)
}

func receiveSelectedTip(incomingRoute *router.Route) (selectedTipHash *daghash.Hash, err error) {
	message, err := incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgSelectedTip := message.(*wire.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, nil
}

// HandleGetSelectedTip handles getSelectedTip messages
func HandleGetSelectedTip(incomingRoute *router.Route, outgoingRoute *router.Route, dag *blockdag.BlockDAG) error {
	for {
		err := receiveGetSelectedTip(incomingRoute)
		if err != nil {
			return err
		}

		selectedTipHash := dag.SelectedTipHash()
		err = sendSelectedTipHash(outgoingRoute, selectedTipHash)
		if err != nil {
			return err
		}
	}
}

func receiveGetSelectedTip(incomingRoute *router.Route) error {
	message, err := incomingRoute.Dequeue()
	if err != nil {
		return err
	}
	_, ok := message.(*wire.MsgGetSelectedTip)
	if !ok {
		panic(errors.Errorf("received unexpected message type. "+
			"expected: %s, got: %s", wire.CmdGetSelectedTip, message.Command()))
	}

	return nil
}

func sendSelectedTipHash(outgoingRoute *router.Route, selectedTipHash *daghash.Hash) error {
	msgSelectedTip := wire.NewMsgSelectedTip(selectedTipHash)
	return outgoingRoute.Enqueue(msgSelectedTip)
}
