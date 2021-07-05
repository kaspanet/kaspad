package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestNextIBDBlocks) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestNextHeaders is nil")
	}
	return &appmessage.MsgRequestNextIBDBlocks{}, nil
}

func (x *KaspadMessage_RequestNextIBDBlocks) fromAppMessage(_ *appmessage.MsgRequestNextIBDBlocks) error {
	return nil
}
