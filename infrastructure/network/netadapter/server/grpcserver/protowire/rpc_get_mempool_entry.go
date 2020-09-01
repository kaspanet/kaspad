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
	var err *appmessage.RPCError
	if x.GetMempoolEntryResponse.Error != nil {
		err = &appmessage.RPCError{Message: x.GetMempoolEntryResponse.Error.Message}
	}
	return &appmessage.GetMempoolEntryResponseMessage{
		Error: err,
	}, nil
}

func (x *KaspadMessage_GetMempoolEntryResponse) fromAppMessage(message *appmessage.GetMempoolEntryResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.GetMempoolEntryResponse = &GetMempoolEntryResponseMessage{
		Error: err,
	}
	return nil
}
