package protowire

import (
	"github.com/kaspanet/kaspad/domainmessage"
	"github.com/pkg/errors"
)

func (x *KaspadMessage_RequestTransactions) toDomainMessage() (domainmessage.Message, error) {
	if len(x.RequestTransactions.Ids) > domainmessage.MaxInvPerRequestTransactionsMsg {
		return nil, errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestTransactions.Ids), domainmessage.MaxInvPerRequestTransactionsMsg)
	}

	ids, err := protoTransactionIDsToWire(x.RequestTransactions.Ids)
	if err != nil {
		return nil, err
	}
	return &domainmessage.MsgRequestTransactions{IDs: ids}, nil
}

func (x *KaspadMessage_RequestTransactions) fromDomainMessage(msgGetTransactions *domainmessage.MsgRequestTransactions) error {
	if len(msgGetTransactions.IDs) > domainmessage.MaxInvPerRequestTransactionsMsg {
		return errors.Errorf("too many hashes for message "+
			"[count %d, max %d]", len(x.RequestTransactions.Ids), domainmessage.MaxInvPerRequestTransactionsMsg)
	}

	x.RequestTransactions = &RequestTransactionsMessage{
		Ids: wireTransactionIDsToProto(msgGetTransactions.IDs),
	}
	return nil
}
