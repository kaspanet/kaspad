package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *KaspadMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}
