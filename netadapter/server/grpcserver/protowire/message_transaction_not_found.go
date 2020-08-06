package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
)

func (x *KaspadMessage_TransactionNotFound) toWireMessage() (domainmessage.Message, error) {
	id, err := x.TransactionNotFound.Id.toWire()
	if err != nil {
		return nil, err
	}
	return domainmessage.NewMsgTransactionNotFound(id), nil
}

func (x *KaspadMessage_TransactionNotFound) fromWireMessage(msgTransactionsNotFound *domainmessage.MsgTransactionNotFound) error {
	x.TransactionNotFound = &TransactionNotFoundMessage{
		Id: wireTransactionIDToProto(msgTransactionsNotFound.ID),
	}
	return nil
}
