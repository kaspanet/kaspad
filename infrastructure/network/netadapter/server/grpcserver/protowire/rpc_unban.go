package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_UnbanRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.UnbanRequestMessage{
		IP: x.UnbanRequest.Ip,
	}, nil
}

func (x *KaspadMessage_UnbanRequest) fromAppMessage(message *appmessage.UnbanRequestMessage) error {
	x.UnbanRequest = &UnbanRequestMessage{Ip: message.IP}
	return nil
}

func (x *KaspadMessage_UnbanResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.UnbanResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.UnbanResponse.Error.Message}
	}
	return &appmessage.UnbanResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_UnbanResponse) fromAppMessage(message *appmessage.UnbanResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.UnbanResponse = &UnbanResponseMessage{
		Error: err,
	}
	return nil
}
