package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetVirtualSelectedParentBlueScoreRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetVirtualSelectedParentBlueScoreRequestMessage{}, nil
}

func (x *KaspadMessage_GetVirtualSelectedParentBlueScoreRequest) fromAppMessage(message *appmessage.GetVirtualSelectedParentBlueScoreRequestMessage) error {
	x.GetVirtualSelectedParentBlueScoreRequest = &GetVirtualSelectedParentBlueScoreRequestMessage{}
	return nil
}

func (x *KaspadMessage_GetVirtualSelectedParentBlueScoreResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetVirtualSelectedParentBlueScoreResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetVirtualSelectedParentBlueScoreResponse.Error.Message}
	}
	return &appmessage.GetVirtualSelectedParentBlueScoreResponseMessage{
		BlueScore: x.GetVirtualSelectedParentBlueScoreResponse.BlueScore,
		Error:     err,
	}, nil
}

func (x *KaspadMessage_GetVirtualSelectedParentBlueScoreResponse) fromAppMessage(message *appmessage.GetVirtualSelectedParentBlueScoreResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetVirtualSelectedParentBlueScoreResponse = &GetVirtualSelectedParentBlueScoreResponseMessage{
		BlueScore: message.BlueScore,
		Error:     err,
	}
	return nil
}
