package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_BanRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.BanRequestMessage{
		IP: x.BanRequest.Ip,
	}, nil
}

func (x *KaspadMessage_BanRequest) fromAppMessage(message *appmessage.BanRequestMessage) error {
	x.BanRequest = &BanRequestMessage{Ip: message.IP}
	return nil
}

func (x *KaspadMessage_BanResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.BanResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.BanResponse.Error.Message}
	}
	return &appmessage.BanResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_BanResponse) fromAppMessage(message *appmessage.BanResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.BanResponse = &BanResponseMessage{
		Error: err,
	}
	return nil
}
