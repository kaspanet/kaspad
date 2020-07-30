package selectedtip

import (
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/netadapter/router"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

// GetSelectedTipContext is the interface for the context needed for the HandleGetSelectedTip flow.
type GetSelectedTipContext interface {
	DAG() *blockdag.BlockDAG
}

type handleGetSelectedTipFlow struct {
	GetSelectedTipContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleGetSelectedTip handles getSelectedTip messages
func HandleGetSelectedTip(context GetSelectedTipContext, incomingRoute *router.Route, outgoingRoute *router.Route) error {
	flow := &handleGetSelectedTipFlow{
		GetSelectedTipContext: context,
		incomingRoute:         incomingRoute,
		outgoingRoute:         outgoingRoute,
	}
	return flow.start()
}

func (flow *handleGetSelectedTipFlow) start() error {
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

func (flow *handleGetSelectedTipFlow) receiveGetSelectedTip() error {
	message, err := flow.incomingRoute.Dequeue()
	if err != nil {
		return err
	}
	_, ok := message.(*wire.MsgRequestSelectedTip)
	if !ok {
		return errors.Errorf("received unexpected message type. "+
			"expected: %s, got: %s", wire.CmdRequestSelectedTip, message.Command())
	}

	return nil
}

func (flow *handleGetSelectedTipFlow) sendSelectedTipHash() error {
	msgSelectedTip := wire.NewMsgSelectedTip(flow.DAG().SelectedTipHash())
	return flow.outgoingRoute.Enqueue(msgSelectedTip)
}
