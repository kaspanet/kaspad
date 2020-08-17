package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_Reject) toWireMessage() (appmessage.Message, error) {
	return &appmessage.MsgReject{
		Reason: x.Reject.Reason,
	}, nil
}

func (x *KaspadMessage_Reject) fromWireMessage(msgReject *appmessage.MsgReject) error {
	x.Reject = &RejectMessage{
		Reason: msgReject.Reason,
	}
	return nil
}
