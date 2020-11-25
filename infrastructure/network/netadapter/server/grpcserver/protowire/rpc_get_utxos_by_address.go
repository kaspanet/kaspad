package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetUTXOsByAddressRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetUTXOsByAddressRequestMessage{
		Address: x.GetUTXOsByAddressRequest.Address,
	}, nil
}

func (x *KaspadMessage_GetUTXOsByAddressRequest) fromAppMessage(message *appmessage.GetUTXOsByAddressRequestMessage) error {
	x.GetUTXOsByAddressRequest = &GetUTXOsByAddressRequestMessage{
		Address: message.Address,
	}
	return nil
}

func (x *KaspadMessage_GetUTXOsByAddressResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetUTXOsByAddressResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetUTXOsByAddressResponse.Error.Message}
	}

	utxosVerboseData := make([]*appmessage.UTXOVerboseData, len(x.GetUTXOsByAddressResponse.UtxosVerboseData))
	for i, utxoVerboseData := range x.GetUTXOsByAddressResponse.UtxosVerboseData {
		appUTXOVerboseData, err := utxoVerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		utxosVerboseData[i] = appUTXOVerboseData
	}

	return &appmessage.GetUTXOsByAddressResponseMessage{
		Address:          x.GetUTXOsByAddressResponse.Address,
		UTXOsVerboseData: utxosVerboseData,
		Error:            err,
	}, nil
}

func (x *KaspadMessage_GetUTXOsByAddressResponse) fromAppMessage(message *appmessage.GetUTXOsByAddressResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}

	utxosVerboseData := make([]*UTXOVerboseData, len(message.UTXOsVerboseData))
	for i, utxoVerboseData := range message.UTXOsVerboseData {
		protoUTXOVerboseData := &UTXOVerboseData{}
		err := protoUTXOVerboseData.fromAppMessage(utxoVerboseData)
		if err != nil {
			return err
		}
		utxosVerboseData[i] = protoUTXOVerboseData
	}

	x.GetUTXOsByAddressResponse = &GetUTXOsByAddressResponseMessage{
		Address:          message.Address,
		UtxosVerboseData: utxosVerboseData,
		Error:            err,
	}
	return nil
}

func (x *UTXOVerboseData) toAppMessage() (*appmessage.UTXOVerboseData, error) {
	return &appmessage.UTXOVerboseData{
		Amount:         x.Amount,
		ScriptPubKey:   x.ScriptPubKey,
		BlockBlueScore: x.BlockBlueScore,
		IsCoinbase:     x.IsCoinbase,
		TxID:           x.TxID,
		Index:          x.Index,
	}, nil
}

func (x *UTXOVerboseData) fromAppMessage(message *appmessage.UTXOVerboseData) error {
	*x = UTXOVerboseData{
		Amount:         message.Amount,
		ScriptPubKey:   message.ScriptPubKey,
		BlockBlueScore: message.BlockBlueScore,
		IsCoinbase:     message.IsCoinbase,
		TxID:           message.TxID,
		Index:          message.Index,
	}
	return nil
}
