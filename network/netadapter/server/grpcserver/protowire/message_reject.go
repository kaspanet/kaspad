package protowire

import (
	"github.com/kaspanet/kaspad/network/domainmessage"
)

func (x *KaspadMessage_Reject) toWireMessage() (domainmessage.Message, error) {
	return &domainmessage.MsgReject{
		Reason: x.Reject.Reason,
	}, nil
}

func (x *KaspadMessage_Reject) fromWireMessage(msgReject *domainmessage.MsgReject) error {
	x.Reject = &RejectMessage{
		Reason: msgReject.Reason,
	}
	return nil
}
