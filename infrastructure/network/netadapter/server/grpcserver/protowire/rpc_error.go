package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_RpcError) toAppMessage() (appmessage.Message, error) {
	return &appmessage.RPCErrorMessage{
		Message: x.RpcError.Message,
	}, nil
}

func (x *KaspadMessage_RpcError) fromAppMessage(message *appmessage.RPCErrorMessage) error {
	x.RpcError = &RPCErrorMessage{
		Message: message.Message,
	}
	return nil
}
