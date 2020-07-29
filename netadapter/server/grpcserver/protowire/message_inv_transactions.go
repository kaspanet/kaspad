package protowire

import (
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_InvTransactions) toWireMessage() (wire.Message, error) {
	if len(x.InvTransactions.Ids) > wire.MaxInvPerTxInvMsg {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.InvTransactions.Ids), wire.MaxInvPerTxInvMsg)
	}

	ids, err := protoTransactionIDsToWire(x.InvTransactions.Ids)
	if err != nil {
		return nil, err
	}
	return &wire.MsgInvTransaction{TxIDs: ids}, nil
}

func (x *KaspadMessage_InvTransactions) fromWireMessage(msgInvTransaction *wire.MsgInvTransaction) error {
	if len(msgInvTransaction.TxIDs) > wire.MaxInvPerTxInvMsg {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgInvTransaction.TxIDs), wire.MaxInvPerTxInvMsg)
	}

	x.InvTransactions = &InvTransactionsMessage{
		Ids: wireTransactionIDsToProto(msgInvTransaction.TxIDs),
	}
	return nil
}
