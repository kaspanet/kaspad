package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_DoneHeaders) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_DoneHeaders is nil")
	}
	return &appmessage.MsgDoneIBDBlocks{}, nil
}

func (x *KaspadMessage_DoneHeaders) fromAppMessage(_ *appmessage.MsgDoneIBDBlocks) error {
	return nil
}
