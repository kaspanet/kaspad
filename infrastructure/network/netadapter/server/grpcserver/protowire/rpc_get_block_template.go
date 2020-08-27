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
	transactions := make([]appmessage.GetBlockTemplateTransactionMessage, len(x.GetBlockTemplateResponse.Transactions))
	for i, transaction := range x.GetBlockTemplateResponse.Transactions {
		appMessageTransaction, err := transaction.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactions[i] = *appMessageTransaction.(*appmessage.GetBlockTemplateTransactionMessage)
	}
	return &appmessage.GetBlockTemplateResponseMessage{
		Bits:                 x.GetBlockTemplateResponse.Bits,
		CurrentTime:          x.GetBlockTemplateResponse.CurrentTime,
		ParentHashes:         x.GetBlockTemplateResponse.ParentHashes,
		MassLimit:            int(x.GetBlockTemplateResponse.MassLimit),
		Transactions:         transactions,
		HashMerkleRoot:       x.GetBlockTemplateResponse.HashMerkleRoot,
		AcceptedIDMerkleRoot: x.GetBlockTemplateResponse.AcceptedIDMerkleRoot,
		UTXOCommitment:       x.GetBlockTemplateResponse.UTXOCommitment,
		Version:              x.GetBlockTemplateResponse.Version,
		LongPollID:           x.GetBlockTemplateResponse.LongPollID,
		TargetDifficulty:     x.GetBlockTemplateResponse.TargetDifficulty,
		MinTime:              x.GetBlockTemplateResponse.MinTime,
		MaxTime:              x.GetBlockTemplateResponse.MaxTime,
		MutableFields:        x.GetBlockTemplateResponse.MutableFields,
		NonceRange:           x.GetBlockTemplateResponse.NonceRange,
		IsSynced:             x.GetBlockTemplateResponse.IsSynced,
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
	x.GetBlockTemplateResponse = &GetBlockTemplateResponseMessage{
		Bits:                 message.Bits,
		CurrentTime:          message.CurrentTime,
		ParentHashes:         message.ParentHashes,
		MassLimit:            int32(message.MassLimit),
		Transactions:         transactions,
		HashMerkleRoot:       message.HashMerkleRoot,
		AcceptedIDMerkleRoot: message.AcceptedIDMerkleRoot,
		UTXOCommitment:       message.UTXOCommitment,
		Version:              message.Version,
		LongPollID:           message.LongPollID,
		TargetDifficulty:     message.TargetDifficulty,
		MinTime:              message.MinTime,
		MaxTime:              message.MaxTime,
		MutableFields:        message.MutableFields,
		NonceRange:           message.NonceRange,
		IsSynced:             message.IsSynced,
	}
	return nil
}

func (x *GetBlockTemplateTransactionMessage) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetBlockTemplateTransactionMessage{
		Data:    x.Data,
		ID:      x.ID,
		Depends: x.Depends,
		Mass:    x.Mass,
		Fee:     x.Fee,
	}, nil
}

func (x *GetBlockTemplateTransactionMessage) fromAppMessage(message *appmessage.GetBlockTemplateTransactionMessage) error {
	x.Data = message.Data
	x.ID = message.ID
	x.Depends = message.Depends
	x.Mass = message.Mass
	x.Fee = message.Fee
	return nil
}
