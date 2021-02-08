package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetInfoRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetInfoRequestMessage{}, nil
}

func (x *KaspadMessage_GetInfoRequest) fromAppMessage(_ *appmessage.GetInfoRequestMessage) error {
	x.GetInfoRequest = &GetInfoRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetInfoResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetInfoResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetInfoResponse.Error.Message}
	}
	return &appmessage.GetInfoResponseMessage{
		P2PID: x.GetInfoResponse.P2PId,
		Error: err,
	}, nil
}

func (x *KaspadMessage_GetInfoResponse) fromAppMessage(message *appmessage.GetInfoResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetInfoResponse = &GetInfoResponseMessage{
		P2PId: message.P2PID,
		Error: err,
	}
	return nil
}
