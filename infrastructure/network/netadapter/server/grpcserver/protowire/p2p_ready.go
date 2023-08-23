package protowire

import (
	"github.com/c4ei/yunseokyeol/app/appmessage"
	"github.com/pkg/errors"
)

func (x *C4exdMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "C4exdMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *C4exdMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}
