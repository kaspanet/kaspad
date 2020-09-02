package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_SendRawTransactionRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.SendRawTransactionRequestMessage{
		TransactionHex: x.SendRawTransactionRequest.TransactionHex,
	}, nil
}

func (x *KaspadMessage_SendRawTransactionRequest) fromAppMessage(message *appmessage.SendRawTransactionRequestMessage) error {
	x.SendRawTransactionRequest = &SendRawTransactionRequestMessage{
		TransactionHex: message.TransactionHex,
	}
	return nil
}

func (x *KaspadMessage_SendRawTransactionResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.SendRawTransactionResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.SendRawTransactionResponse.Error.Message}
	}
	return &appmessage.SendRawTransactionResponseMessage{
		TxID:  x.SendRawTransactionResponse.TxId,
		Error: err,
	}, nil
}

func (x *KaspadMessage_SendRawTransactionResponse) fromAppMessage(message *appmessage.SendRawTransactionResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.SendRawTransactionResponse = &SendRawTransactionResponseMessage{
		TxId:  message.TxID,
		Error: err,
	}
	return nil
}
