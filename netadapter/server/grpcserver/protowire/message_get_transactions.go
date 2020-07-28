package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_GetTransactions) toWireMessage() (*wire.MsgGetTransactions, error) {
	if len(x.GetTransactions.Ids) > wire.MaxInvPerGetTransactionsMsg {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.GetTransactions.Ids), wire.MaxInvPerGetTransactionsMsg)
	}

	ids, err := protoTransactionIDsToWire(x.GetTransactions.Ids)
	if err != nil {
		return nil, err
	}
	return &wire.MsgGetTransactions{IDs: ids}, nil
}

func (x *KaspadMessage_GetTransactions) fromWireMessage(msgGetTransactions *wire.MsgGetTransactions) error {
	if len(x.GetTransactions.Ids) > wire.MaxInvPerGetTransactionsMsg {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.GetTransactions.Ids), wire.MaxInvPerGetTransactionsMsg)
	}

	x.GetTransactions = &GetTransactionsMessage{
		Ids: wireTransactionIDsToProto(msgGetTransactions.IDs),
	}
	return nil
}
