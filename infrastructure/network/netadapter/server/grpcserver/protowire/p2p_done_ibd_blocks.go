package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_DoneIbdBlocks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_DoneIBDBlocks is nil")
	}
	return &appmessage.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneIbdBlocks) fromAppMessage(_ *appmessage.MsgDoneIBDBlocks) error {
	return nil
}
