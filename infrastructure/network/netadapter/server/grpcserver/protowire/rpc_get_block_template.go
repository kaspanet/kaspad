package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

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
	msgBlock, err := x.GetBlockTemplateResponse.BlockMessage.toAppMessage()
	if err != nil {
		return nil, err
	}
	return appmessage.NewGetBlockTemplateResponseMessage(msgBlock.(*appmessage.MsgBlock)), nil
}

func (x *KaspadMessage_GetBlockTemplateResponse) fromAppMessage(message *appmessage.GetBlockTemplateResponseMessage) error {
	x.GetBlockTemplateResponse = &GetBlockTemplateResponseMessage{
		BlockMessage: &BlockMessage{},
	}
	return x.GetBlockTemplateResponse.BlockMessage.fromAppMessage(message.MsgBlock)
}
