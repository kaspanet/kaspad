package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_SubmitTransactionRequest) toAppMessage() (appmessage.Message, error) {
	msgTx, err := x.SubmitTransactionRequest.Transaction.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.SubmitTransactionRequestMessage{
		Transaction: msgTx.(*appmessage.MsgTx),
	}, nil
}

func (x *KaspadMessage_SubmitTransactionRequest) fromAppMessage(message *appmessage.SubmitTransactionRequestMessage) error {
	x.SubmitTransactionRequest = &SubmitTransactionRequestMessage{
		Transaction: &TransactionMessage{},
	}
	x.SubmitTransactionRequest.Transaction.fromAppMessage(message.Transaction)
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
