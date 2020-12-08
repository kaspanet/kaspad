package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetBlockCountRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockCountRequestMessage{}, nil
}

func (x *KaspadMessage_GetBlockCountRequest) fromAppMessage(_ *appmessage.GetBlockCountRequestMessage) error {
	x.GetBlockCountRequest = &GetBlockCountRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetBlockCountResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetBlockCountResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockCountResponse.Error.Message}
	}
	return &appmessage.GetBlockCountResponseMessage{
		BlockCount:  x.GetBlockCountResponse.BlockCount,
		HeaderCount: x.GetBlockCountResponse.HeaderCount,
		Error:       err,
	}, nil
}

func (x *KaspadMessage_GetBlockCountResponse) fromAppMessage(message *appmessage.GetBlockCountResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlockCountResponse = &GetBlockCountResponseMessage{
		BlockCount:  message.BlockCount,
		HeaderCount: message.HeaderCount,
		Error:       err,
	}
	return nil
}
