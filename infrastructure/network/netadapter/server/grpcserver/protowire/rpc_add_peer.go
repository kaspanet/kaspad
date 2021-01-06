package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_AddPeerRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.AddPeerRequestMessage{
		Address:     x.AddPeerRequest.Address,
		IsPermanent: x.AddPeerRequest.IsPermanent,
	}, nil
}

func (x *KaspadMessage_AddPeerRequest) fromAppMessage(message *appmessage.AddPeerRequestMessage) error {
	x.AddPeerRequest = &AddPeerRequestMessage{
		Address:     message.Address,
		IsPermanent: message.IsPermanent,
	}
	return nil
}

func (x *KaspadMessage_AddPeerResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.AddPeerResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.AddPeerResponse.Error.Message}
	}
	return &appmessage.AddPeerResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_AddPeerResponse) fromAppMessage(message *appmessage.AddPeerResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.AddPeerResponse = &AddPeerResponseMessage{
		Error: err,
	}
	return nil
}
