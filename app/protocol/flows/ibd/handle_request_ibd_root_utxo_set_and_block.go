package ibd

import (
	"errors"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleRequestIBDRootUTXOSetAndBlockContext is the interface for the context needed for the HandleRequestIBDRootUTXOSetAndBlock flow.
type HandleRequestIBDRootUTXOSetAndBlockContext interface {
	Domain() domain.Domain
}

type handleRequestIBDRootUTXOSetAndBlockFlow struct {
	HandleRequestIBDRootUTXOSetAndBlockContext
	incomingRoute, outgoingRoute *router.Route
}

// HandleRequestIBDRootUTXOSetAndBlock listens to appmessage.MsgRequestIBDRootUTXOSetAndBlock messages and sends
// the IBD root UTXO set and block body.
func HandleRequestIBDRootUTXOSetAndBlock(context HandleRequestIBDRootUTXOSetAndBlockContext, incomingRoute,
	outgoingRoute *router.Route) error {
	flow := &handleRequestIBDRootUTXOSetAndBlockFlow{
		HandleRequestIBDRootUTXOSetAndBlockContext: context,
		incomingRoute: incomingRoute,
		outgoingRoute: outgoingRoute,
	}

	return flow.start()
}

func (flow *handleRequestIBDRootUTXOSetAndBlockFlow) start() error {
	for {
		message, err := flow.incomingRoute.Dequeue()
		if err != nil {
			return err
		}
		msgRequestIBDRootUTXOSetAndBlock := message.(*appmessage.MsgRequestIBDRootUTXOSetAndBlock)

		utxoSet, err := flow.Domain().Consensus().GetPruningPointUTXOSet(msgRequestIBDRootUTXOSetAndBlock.IBDRoot)
		if err != nil {
			if errors.Is(err, ruleerrors.ErrWrongPruningPointHash) {
				err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDRootNotFound())
				if err != nil {
					return err
				}

				continue
			}
		}

		block, err := flow.Domain().Consensus().GetBlock(msgRequestIBDRootUTXOSetAndBlock.IBDRoot)
		if err != nil {
			return err
		}

		err = flow.outgoingRoute.Enqueue(appmessage.NewMsgIBDRootUTXOSetAndBlock(utxoSet,
			appmessage.DomainBlockToMsgBlock(block)))
		if err != nil {
			return err
		}
	}
}
