package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_NotifyTransactionAddedRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.NotifyTransactionAddedRequestMessage{
		Addresses: x.NotifyTransactionAddedRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_NotifyTransactionAddedRequest) fromAppMessage(message *appmessage.NotifyTransactionAddedRequestMessage) error {
	x.NotifyTransactionAddedRequest = &NotifyTransactionAddedRequestMessage{
		Addresses: message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_NotifyTransactionAddedResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.NotifyTransactionAddedResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.NotifyTransactionAddedResponse.Error.Message}
	}
	return &appmessage.NotifyTransactionAddedResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_NotifyTransactionAddedResponse) fromAppMessage(message *appmessage.NotifyTransactionAddedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyTransactionAddedResponse = &NotifyTransactionAddedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *KaspadMessage_TransactionAddedNotification) toAppMessage() (appmessage.Message, error) {
	utxosVerboseData := make([]*appmessage.UTXOVerboseData, len(x.TransactionAddedNotification.UtxosVerboseData))
	for i, utxoVerboseData := range x.TransactionAddedNotification.UtxosVerboseData {
		appUTXOVerboseData, err := utxoVerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		utxosVerboseData[i] = appUTXOVerboseData
	}

	msgTx, err := x.TransactionAddedNotification.Transaction.toAppMessage()
	if err != nil {
		return nil, err
	}

	return &appmessage.TransactionAddedNotificationMessage{
		Addresses:        x.TransactionAddedNotification.Addresses,
		BlockHash:        x.TransactionAddedNotification.BlockHash,
		UTXOsVerboseData: utxosVerboseData,
		Transaction:      msgTx.(*appmessage.MsgTx),
		Status:           x.TransactionAddedNotification.Status,
	}, nil
}

func (x *KaspadMessage_TransactionAddedNotification) fromAppMessage(message *appmessage.TransactionAddedNotificationMessage) error {
	utxosVerboseData := make([]*UTXOVerboseData, len(message.UTXOsVerboseData))
	for i, utxoVerboseData := range message.UTXOsVerboseData {
		protoUTXOVerboseData := &UTXOVerboseData{}
		err := protoUTXOVerboseData.fromAppMessage(utxoVerboseData)
		if err != nil {
			return err
		}
		utxosVerboseData[i] = protoUTXOVerboseData
	}
	protoTx := new(TransactionMessage)
	protoTx.fromAppMessage(message.Transaction)

	x.TransactionAddedNotification = &TransactionAddedNotificationMessage{
		Addresses:        message.Addresses,
		BlockHash:        message.BlockHash,
		UtxosVerboseData: utxosVerboseData,
		Transaction:      protoTx,
		Status:           message.Status,
	}

	return nil
}
