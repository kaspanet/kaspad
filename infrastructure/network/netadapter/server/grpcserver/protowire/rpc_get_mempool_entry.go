package protowire

import "github.com/kaspanet/kaspad/app/appmessage"

func (x *KaspadMessage_GetMempoolEntryRequest) toAppMessage() (appmessage.Message, error) {
	return &appmessage.GetMempoolEntryRequestMessage{
		TxID: x.GetMempoolEntryRequest.TxId,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntryRequest) fromAppMessage(message *appmessage.GetMempoolEntryRequestMessage) error {
	x.GetMempoolEntryRequest = &GetMempoolEntryRequestMessage{
		TxId: message.TxID,
	}
	return nil
}

func (x *KaspadMessage_GetMempoolEntryResponse) toAppMessage() (appmessage.Message, error) {
	var rpcErr *appmessage.RPCError
	if x.GetMempoolEntryResponse.Error != nil {
		rpcErr = &appmessage.RPCError{Message: x.GetMempoolEntryResponse.Error.Message}
	}
	transactionVerboseData, err := x.GetMempoolEntryResponse.TransactionVerboseData.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.GetMempoolEntryResponseMessage{
		Fee:                    x.GetMempoolEntryResponse.Fee,
		TransactionVerboseData: transactionVerboseData,
		Error:                  rpcErr,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntryResponse) fromAppMessage(message *appmessage.GetMempoolEntryResponseMessage) error {
	var rpcErr *RPCError
	if message.Error != nil {
		rpcErr = &RPCError{Message: message.Error.Message}
	}
	transactionVerboseData := &TransactionVerboseData{}
	err := transactionVerboseData.fromAppMessage(message.TransactionVerboseData)
	if err != nil {
		return err
	}
	x.GetMempoolEntryResponse = &GetMempoolEntryResponseMessage{
		Fee:                    message.Fee,
		TransactionVerboseData: transactionVerboseData,
		Error:                  rpcErr,
	}
	return nil
}
