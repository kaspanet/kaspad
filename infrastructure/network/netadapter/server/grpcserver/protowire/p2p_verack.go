package protowire

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
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
