package protowire

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_InvTransactions) toAppMessage() (appmessage.Message, error) {
	if len(x.InvTransactions.Ids) > appmessage.MaxInvPerTxInvMsg {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.InvTransactions.Ids), appmessage.MaxInvPerTxInvMsg)
	}

	ids, err := protoTransactionIDsToDomain(x.InvTransactions.Ids)
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgInvTransaction{TxIDs: ids}, nil
}

func (x *KaspadMessage_InvTransactions) fromAppMessage(msgInvTransaction *appmessage.MsgInvTransaction) error {
	if len(msgInvTransaction.TxIDs) > appmessage.MaxInvPerTxInvMsg {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(msgInvTransaction.TxIDs), appmessage.MaxInvPerTxInvMsg)
	}

	x.InvTransactions = &InvTransactionsMessage{
		Ids: wireTransactionIDsToProto(msgInvTransaction.TxIDs),
	}
	return nil
}
