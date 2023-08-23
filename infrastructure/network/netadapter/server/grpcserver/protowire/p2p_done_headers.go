package protowire

import (
	"github.com/c4ei/yunseokyeol/app/appmessage"
	"github.com/pkg/errors"
)

func (x *C4exdMessage_DoneHeaders) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "C4exdMessage_DoneHeaders is nil")
	}
	return &appmessage.MsgDoneHeaders{}, nil
}

func (x *C4exdMessage_DoneHeaders) fromAppMessage(_ *appmessage.MsgDoneHeaders) error {
	return nil
}
