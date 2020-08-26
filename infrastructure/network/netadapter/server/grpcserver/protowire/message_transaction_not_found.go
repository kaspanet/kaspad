package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_TransactionNotFound) toAppMessage() (appmessage.Message, error) {
	id, err := x.TransactionNotFound.Id.toWire()
	if err != nil {
		return nil, err
	}
	return appmessage.NewMsgTransactionNotFound(id), nil
}

func (x *KaspadMessage_TransactionNotFound) fromAppMessage(msgTransactionsNotFound *appmessage.MsgTransactionNotFound) error {
	x.TransactionNotFound = &TransactionNotFoundMessage{
		Id: wireTransactionIDToProto(msgTransactionsNotFound.ID),
	}
	return nil
}
