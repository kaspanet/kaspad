package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetBlockTemplateRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockTemplateRequestMessage{
		PayAddress: x.GetBlockTemplateRequest.PayAddress,
		LongPollID: x.GetBlockTemplateRequest.LongPollId,
	}, nil
}

func (x *KaspadMessage_GetBlockTemplateRequest) fromAppMessage(message *appmessage.GetBlockTemplateRequestMessage) error {
	x.GetBlockTemplateRequest = &GetBlockTemplateRequestMessage{
		PayAddress: message.PayAddress,
		LongPollId: message.LongPollID,
	}
	return nil
}

func (x *KaspadMessage_GetBlockTemplateResponse) toAppMessage() (appmessage.Message, error) {
	transactions := make([]appmessage.GetBlockTemplateTransactionMessage, len(x.GetBlockTemplateResponse.Transactions))
	for i, transaction := range x.GetBlockTemplateResponse.Transactions {
		appMessageTransaction, err := transaction.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactions[i] = *appMessageTransaction.(*appmessage.GetBlockTemplateTransactionMessage)
	}
	var err *appmessage.RPCError
	if x.GetBlockTemplateResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetBlockTemplateResponse.Error.Message}
	}
	return &appmessage.GetBlockTemplateResponseMessage{
		Bits:                 x.GetBlockTemplateResponse.Bits,
		CurrentTime:          x.GetBlockTemplateResponse.CurrentTime,
		ParentHashes:         x.GetBlockTemplateResponse.ParentHashes,
		MassLimit:            int(x.GetBlockTemplateResponse.MassLimit),
		Transactions:         transactions,
		HashMerkleRoot:       x.GetBlockTemplateResponse.HashMerkleRoot,
		AcceptedIDMerkleRoot: x.GetBlockTemplateResponse.AcceptedIDMerkleRoot,
		UTXOCommitment:       x.GetBlockTemplateResponse.UtxoCommitment,
		Version:              x.GetBlockTemplateResponse.Version,
		LongPollID:           x.GetBlockTemplateResponse.LongPollId,
		TargetDifficulty:     x.GetBlockTemplateResponse.TargetDifficulty,
		MinTime:              x.GetBlockTemplateResponse.MinTime,
		MaxTime:              x.GetBlockTemplateResponse.MaxTime,
		MutableFields:        x.GetBlockTemplateResponse.MutableFields,
		NonceRange:           x.GetBlockTemplateResponse.NonceRange,
		IsSynced:             x.GetBlockTemplateResponse.IsSynced,
		IsConnected:          x.GetBlockTemplateResponse.IsConnected,
		Error:                err,
	}, nil
}

func (x *KaspadMessage_GetBlockTemplateResponse) fromAppMessage(message *appmessage.GetBlockTemplateResponseMessage) error {
	transactions := make([]*GetBlockTemplateTransactionMessage, len(message.Transactions))
	for i, transaction := range message.Transactions {
		protoMessageTransaction := GetBlockTemplateTransactionMessage{}
		err := protoMessageTransaction.fromAppMessage(&transaction)
		if err != nil {
			return err
		}
		transactions[i] = &protoMessageTransaction
	}
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetBlockTemplateResponse = &GetBlockTemplateResponseMessage{
		Bits:                 message.Bits,
		CurrentTime:          message.CurrentTime,
		ParentHashes:         message.ParentHashes,
		MassLimit:            int32(message.MassLimit),
		Transactions:         transactions,
		HashMerkleRoot:       message.HashMerkleRoot,
		AcceptedIDMerkleRoot: message.AcceptedIDMerkleRoot,
		UtxoCommitment:       message.UTXOCommitment,
		Version:              message.Version,
		LongPollId:           message.LongPollID,
		TargetDifficulty:     message.TargetDifficulty,
		MinTime:              message.MinTime,
		MaxTime:              message.MaxTime,
		MutableFields:        message.MutableFields,
		NonceRange:           message.NonceRange,
		IsSynced:             message.IsSynced,
		IsConnected:          message.IsConnected,
		Error:                err,
	}
	return nil
}

func (x *GetBlockTemplateTransactionMessage) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockTemplateTransactionMessage{
		Data:    x.Data,
		ID:      x.Id,
		Depends: x.Depends,
		Mass:    x.Mass,
		Fee:     x.Fee,
	}, nil
}

func (x *GetBlockTemplateTransactionMessage) fromAppMessage(message *appmessage.GetBlockTemplateTransactionMessage) error {
	x.Data = message.Data
	x.Id = message.ID
	x.Depends = message.Depends
	x.Mass = message.Mass
	x.Fee = message.Fee
	return nil
}
