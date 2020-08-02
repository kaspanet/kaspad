package protowire

import (
	"github.com/kaspanet/kaspad/wire"
)

func (x *KaspadMessage_TransactionNotFound) toWireMessage() (wire.Message, error) {
	id, err := x.TransactionNotFound.Id.toWire()
	if err != nil {
		return nil, err
	}
	return wire.NewMsgTransactionNotFound(id), nil
}

func (x *KaspadMessage_TransactionNotFound) fromWireMessage(msgTransactionsNotFound *wire.MsgTransactionNotFound) error {
	x.TransactionNotFound = &TransactionNotFoundMessage{
		Id: wireTransactionIDToProto(msgTransactionsNotFound.ID),
	}
	return nil
}
