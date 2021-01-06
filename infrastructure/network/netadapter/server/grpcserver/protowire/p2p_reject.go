package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_Reject) toAppMessage() (appmessage.Message, error) {
	return &appmessage.MsgReject{
		Reason: x.Reject.Reason,
	}, nil
}

func (x *KaspadMessage_Reject) fromAppMessage(msgReject *appmessage.MsgReject) error {
	x.Reject = &RejectMessage{
		Reason: msgReject.Reason,
	}
	return nil
}
