package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_SubmitTransactionRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.SubmitTransactionRequestMessage{
		TransactionHex: x.SubmitTransactionRequest.TransactionHex,
	}, nil
}

func (x *KaspadMessage_SubmitTransactionRequest) fromAppMessage(message *appmessage.SubmitTransactionRequestMessage) error {
	x.SubmitTransactionRequest = &SubmitTransactionRequestMessage{
		TransactionHex: message.TransactionHex,
	}
	return nil
}

func (x *KaspadMessage_SubmitTransactionResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.SubmitTransactionResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.SubmitTransactionResponse.Error.Message}
	}
	return &appmessage.SubmitTransactionResponseMessage{
		TxID:  x.SubmitTransactionResponse.TxId,
		Error: err,
	}, nil
}

func (x *KaspadMessage_SubmitTransactionResponse) fromAppMessage(message *appmessage.SubmitTransactionResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.SubmitTransactionResponse = &SubmitTransactionResponseMessage{
		TxId:  message.TxID,
		Error: err,
	}
	return nil
}
