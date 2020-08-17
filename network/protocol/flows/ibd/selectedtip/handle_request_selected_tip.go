package selectedtip

import (
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/kaspanet/kaspad/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleRequestSelectedTipContext is the interface for the context needed for the HandleRequestSelectedTip flow.
type HandleRequestSelectedTipContext interface {
	DAG() *blockdag.BlockDAG
}

type handleRequestSelectedTipFlow struct {
	HandleRequestSelectedTipContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestSelectedTip handles getSelectedTip messages
func HandleRequestSelectedTip(context HandleRequestSelectedTipContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleRequestSelectedTipFlow{
		HandleRequestSelectedTipContext: context,
		incomingRoute:                   incomingRoute,
		outgoingRoute:                   outgoingRoute,
	}
	return flow.start()
}

func (flow *handleRequestSelectedTipFlow) start() error {
	for {
		err := flow.receiveGetSelectedTip()
		if err != nil {
			return err
		}

		err = flow.sendSelectedTipHash()
		if err != nil {
			return err
		}
	}
}

func (flow *handleRequestSelectedTipFlow) receiveGetSelectedTip() error {
	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return err
	}
	_, ok := message.(*appmessage.MsgRequestSelectedTip)
	if !ok {
		return errors.Errorf("received unexpected message type. "+
			"expected: %s, got: %s", appmessage.CmdRequestSelectedTip, message.Command())
	}

	return nil
}

func (flow *handleRequestSelectedTipFlow) sendSelectedTipHash() error {
	msgSelectedTip := appmessage.NewMsgSelectedTip(flow.DAG().SelectedTipHash())
	return flow.outgoingRoute.Enqueue(msgSelectedTip)
}
