package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetBlockTemplateRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockTemplateRequest is nil")
	}
	return x.GetBlockTemplateRequest.toAppMessage()
}

func (x *KaspadMessage_GetBlockTemplateRequest) fromAppMessage(message *appmessage.GetBlockTemplateRequestMessage) error {
	x.GetBlockTemplateRequest = &GetBlockTemplateRequestMessage{
		PayAddress: message.PayAddress,
	}
	return nil
}

func (x *GetBlockTemplateRequestMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockTemplateRequestMessage is nil")
	}
	return &appmessage.GetBlockTemplateRequestMessage{
		PayAddress: x.PayAddress,
	}, nil
}

func (x *KaspadMessage_GetBlockTemplateResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "KaspadMessage_GetBlockTemplateResponse is nil")
	}
	return x.GetBlockTemplateResponse.toAppMessage()
}

func (x *KaspadMessage_GetBlockTemplateResponse) fromAppMessage(message *appmessage.GetBlockTemplateResponseMessage) error {
	x.GetBlockTemplateResponse = &GetBlockTemplateResponseMessage{
		Block:    &RpcBlock{},
		IsSynced: message.IsSynced,
	}
	return x.GetBlockTemplateResponse.Block.fromAppMessage(message.Block)
}

func (x *GetBlockTemplateResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "GetBlockTemplateResponseMessage is nil")
	}
	msgBlock, err := x.Block.toAppMessage()
	if err != nil {
		return nil, err
	}
	return appmessage.NewGetBlockTemplateResponseMessage(msgBlock, x.IsSynced), nil
}
