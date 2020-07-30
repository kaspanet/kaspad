package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestTransactions) toWireMessage() (wire.Message, error) {
	if len(x.RequestTransactions.Ids) > wire.MaxInvPerRequestTransactionsMsg {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestTransactions.Ids), wire.MaxInvPerRequestTransactionsMsg)
	}

	ids, err := protoTransactionIDsToWire(x.RequestTransactions.Ids)
	if err != nil {
		return nil, err
	}
	return &wire.MsgRequestTransactions{IDs: ids}, nil
}

func (x *KaspadMessage_RequestTransactions) fromWireMessage(msgGetTransactions *wire.MsgRequestTransactions) error {
	if len(x.RequestTransactions.Ids) > wire.MaxInvPerRequestTransactionsMsg {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestTransactions.Ids), wire.MaxInvPerRequestTransactionsMsg)
	}

	x.RequestTransactions = &RequestTransactionsMessage{
		Ids: wireTransactionIDsToProto(msgGetTransactions.IDs),
	}
	return nil
}
