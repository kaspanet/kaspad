package protowire

import (
<<<<<<< Updated upstream
	"github.com/zoomy-network/zoomyd/app/appmessage"
=======
>>>>>>> Stashed changes
	"github.com/pkg/errors"
	"github.com/zoomy-network/zoomyd/app/appmessage"
)

func (x *KaspadMessage_Pong) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_Pong is nil")
	}
	return x.Pong.toAppMessage()
}

func (x *PongMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "PongMessage is nil")
	}
	return &appmessage.MsgPong{
		Nonce: x.Nonce,
	}, nil
}

func (x *KaspadMessage_Pong) fromAppMessage(msgPong *appmessage.MsgPong) error {
	x.Pong = &PongMessage{
		Nonce: msgPong.Nonce,
	}
	return nil
}
