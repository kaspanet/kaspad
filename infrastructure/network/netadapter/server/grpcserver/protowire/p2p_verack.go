package protowire

import (
	"github.com/zoomy-network/zoomyd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *KaspadMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
