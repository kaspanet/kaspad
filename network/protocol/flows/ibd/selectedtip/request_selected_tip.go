package selectedtip

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/network/netadapter/router"
	"github.com/kaspanet/kaspad/network/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/network/protocol/peer"
	"github.com/kaspanet/kaspad/util/daghash"
)

// RequestSelectedTipContext is the interface for the context needed for the RequestSelectedTip flow.
type RequestSelectedTipContext interface {
	DAG() *blockdag.BlockDAG
	StartIBDIfRequired()
}

type requestSelectedTipFlow struct {
	RequestSelectedTipContext
	incomingRoute, outgoingRoute *router.Route
	peer                         *peerpkg.Peer
}

// RequestSelectedTip waits for selected tip requests and handles them
func RequestSelectedTip(context RequestSelectedTipContext, incomingRoute *router.Route,
	outgoingRoute *router.Route, peer *peerpkg.Peer) error {

	flow := &requestSelectedTipFlow{
		RequestSelectedTipContext: context,
		incomingRoute:             incomingRoute,
		outgoingRoute:             outgoingRoute,
		peer:                      peer,
	}
	return flow.start()
}

func (flow *requestSelectedTipFlow) start() error {
	for {
		err := flow.runSelectedTipRequest()
		if err != nil {
			return err
		}
	}
}

func (flow *requestSelectedTipFlow) runSelectedTipRequest() error {

	flow.peer.WaitForSelectedTipRequests()
	defer flow.peer.FinishRequestingSelectedTip()

	err := flow.requestSelectedTip()
	if err != nil {
		return err
	}

	peerSelectedTipHash, err := flow.receiveSelectedTip()
	if err != nil {
		return err
	}

	flow.peer.SetSelectedTipHash(peerSelectedTipHash)
	flow.StartIBDIfRequired()
	return nil
}

func (flow *requestSelectedTipFlow) requestSelectedTip() error {
	msgGetSelectedTip := domainmessage.NewMsgRequestSelectedTip()
	return flow.outgoingRoute.Enqueue(msgGetSelectedTip)
}

func (flow *requestSelectedTipFlow) receiveSelectedTip() (selectedTipHash *daghash.Hash, err error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgSelectedTip := message.(*domainmessage.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, nil
}
