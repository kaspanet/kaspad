package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
)

func (x *KaspadMessage_TransactionNotFound) toAppMessage() (appmessage.Message, error) {
	id, err := x.TransactionNotFound.Id.toDomain()
	if err != nil {
		return nil, err
	}
	return appmessage.NewMsgTransactionNotFound(id), nil
}

func (x *KaspadMessage_TransactionNotFound) fromAppMessage(msgTransactionsNotFound *appmessage.MsgTransactionNotFound) error {
	x.TransactionNotFound = &TransactionNotFoundMessage{
		Id: domainTransactionIDToProto(msgTransactionsNotFound.ID),
	}
	return nil
}
