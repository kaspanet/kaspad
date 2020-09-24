package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetSelectedTipHashRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetSelectedTipHashRequestMessage{}, nil
}

func (x *KaspadMessage_GetSelectedTipHashRequest) fromAppMessage(_ *appmessage.GetSelectedTipHashRequestMessage) error {
	return nil
}

func (x *KaspadMessage_GetSelectedTipHashResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetSelectedTipHashResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetSelectedTipHashResponse.Error.Message}
	}
	return &appmessage.GetSelectedTipHashResponseMessage{
		SelectedTipHash: x.GetSelectedTipHashResponse.SelectedTipHash,
		Error:           err,
	}, nil
}

func (x *KaspadMessage_GetSelectedTipHashResponse) fromAppMessage(message *appmessage.GetSelectedTipHashResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetSelectedTipHashResponse = &GetSelectedTipHashResponseMessage{
		SelectedTipHash: message.SelectedTipHash,
		Error:           err,
	}
	return nil
}
