package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetBlockTemplateRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockTemplateRequestMessage{
		PayAddress: x.GetBlockTemplateRequest.PayAddress,
	}, nil
}

func (x *KaspadMessage_GetBlockTemplateRequest) fromAppMessage(message *appmessage.GetBlockTemplateRequestMessage) error {
	x.GetBlockTemplateRequest = &GetBlockTemplateRequestMessage{
		PayAddress: message.PayAddress,
	}
	return nil
}

func (x *KaspadMessage_GetBlockTemplateResponse) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockTemplateRequestMessage{}, nil
}

func (x *KaspadMessage_GetBlockTemplateResponse) fromAppMessage(_ *appmessage.GetBlockTemplateResponseMessage) error {
	return nil
}
