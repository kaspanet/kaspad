package protowire

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_RequestNextHeaders) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_RequestNextHeaders is nil")
	}
	return &appmessage.MsgRequestNextHeaders{}, nil
}

func (x *KaspadMessage_RequestNextHeaders) fromAppMessage(_ *appmessage.MsgRequestNextHeaders) error {
	return nil
}
