package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_DoneBlocksWithMetaData) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_DoneBlocksWithMetaData is nil")
	}
	return &appmessage.MsgDoneBlocksWithMetaData{}, nil
}

func (x *KaspadMessage_DoneBlocksWithMetaData) fromAppMessage(_ *appmessage.MsgDoneBlocksWithMetaData) error {
	return nil
}
