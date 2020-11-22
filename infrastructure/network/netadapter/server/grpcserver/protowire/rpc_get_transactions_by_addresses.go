package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_GetTransactionsByAddressesRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetTransactionsByAddressesRequestMessage{
		StartingBlockHash: x.GetTransactionsByAddressesRequest.StartingBlockHash,
		Addresses:         x.GetTransactionsByAddressesRequest.Addresses,
	}, nil
}

func (x *KaspadMessage_GetTransactionsByAddressesRequest) fromAppMessage(message *appmessage.GetTransactionsByAddressesRequestMessage) error {
	x.GetTransactionsByAddressesRequest = &GetTransactionsByAddressesRequestMessage{
		StartingBlockHash: x.GetTransactionsByAddressesRequest.StartingBlockHash,
		Addresses:         message.Addresses,
	}
	return nil
}

func (x *KaspadMessage_GetTransactionsByAddressesResponse) toAppMessage() (appmessage.Message, error) {
	var err *appmessage.RPCError
	if x.GetTransactionsByAddressesResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetTransactionsByAddressesResponse.Error.Message}
	}

	transactionsVerboseData := make([]*appmessage.TransactionVerboseData, len(x.GetTransactionsByAddressesResponse.Transactions))
	for i, transactionVerboseData := range x.GetTransactionsByAddressesResponse.Transactions {
		appTransactionVerboseData, err := transactionVerboseData.toAppMessage()
		if err != nil {
			return nil, err
		}
		transactionsVerboseData[i] = appTransactionVerboseData
	}

	return &appmessage.GetTransactionsByAddressesResponseMessage{
		LasBlockScanned: x.GetTransactionsByAddressesResponse.LasBlockScanned,
		Transactions:    transactionsVerboseData,
		Error:           err,
	}, nil
}

func (x *KaspadMessage_GetTransactionsByAddressesResponse) fromAppMessage(message *appmessage.GetTransactionsByAddressesResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}

	transactionsVerboseData := make([]*TransactionVerboseData, len(message.Transactions))
	for i, transactionVerboseData := range message.Transactions {
		protoTransactionVerboseData := &TransactionVerboseData{}
		err := protoTransactionVerboseData.fromAppMessage(transactionVerboseData)
		if err != nil {
			return err
		}
		transactionsVerboseData[i] = protoTransactionVerboseData
	}

	x.GetTransactionsByAddressesResponse = &GetTransactionsByAddressesResponseMessage{
		LasBlockScanned: message.LasBlockScanned,
		Transactions:    transactionsVerboseData,
		Error:           err,
	}
	return nil
}
