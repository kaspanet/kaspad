package selectedtip

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/protocol/common"
	peerpkg "github.com/kaspanet/kaspad/app/protocol/peer"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// RequestSelectedTipContext is the interface for the context needed for the RequestSelectedTip flow.
type RequestSelectedTipContext interface {
	Domain() domain.Domain
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
	msgGetSelectedTip := appmessage.NewMsgRequestSelectedTip()
	return flow.outgoingRoute.Enqueue(msgGetSelectedTip)
}

func (flow *requestSelectedTipFlow) receiveSelectedTip() (selectedTipHash *externalapi.DomainHash, err error) {
	message, err := flow.incomingRoute.DequeueWithTimeout(common.DefaultTimeout)
	if err != nil {
		return nil, err
	}
	msgSelectedTip := message.(*appmessage.MsgSelectedTip)

	return msgSelectedTip.SelectedTipHash, nil
}
