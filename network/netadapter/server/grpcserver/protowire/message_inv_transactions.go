package protowire

import (
	"github.com/kaspanet/kaspad/network/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_InvTransactions) toDomainMessage() (appmessage.Message, error) {
	if len(x.InvTransactions.Ids) > appmessage.MaxInvPerTxInvMsg {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.InvTransactions.Ids), appmessage.MaxInvPerTxInvMsg)
	}

	ids, err := protoTransactionIDsToWire(x.InvTransactions.Ids)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgInvTransaction{TxIDs: ids}, nil
}

func (x *KaspadMessage_InvTransactions) fromDomainMessage(msgInvTransaction *appmessage.MsgInvTransaction) error {
	if len(msgInvTransaction.TxIDs) > appmessage.MaxInvPerTxInvMsg {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgInvTransaction.TxIDs), appmessage.MaxInvPerTxInvMsg)
	}

	x.InvTransactions = &InvTransactionsMessage{
		Ids: wireTransactionIDsToProto(msgInvTransaction.TxIDs),
	}
	return nil
}
