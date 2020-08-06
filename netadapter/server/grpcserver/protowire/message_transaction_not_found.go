package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
)

func (x *KaspadMessage_TransactionNotFound) toDomainMessage() (domainmessage.Message, error) {
	id, err := x.TransactionNotFound.Id.toWire()
	if err != nil {
		return nil, err
	}
	return domainmessage.NewMsgTransactionNotFound(id), nil
}

func (x *KaspadMessage_TransactionNotFound) fromDomainMessage(msgTransactionsNotFound *domainmessage.MsgTransactionNotFound) error {
	x.TransactionNotFound = &TransactionNotFoundMessage{
		Id: wireTransactionIDToProto(msgTransactionsNotFound.ID),
	}
	return nil
}
