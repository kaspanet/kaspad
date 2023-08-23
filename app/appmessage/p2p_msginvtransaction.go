package appmessage

import (
	"github.com/c4ei/yunseokyeol/domain/consensus/model/externalapi"
)

// MaxInvPerTxInvMsg is the maximum number of hashes that can
// be in a single CmdInvTransaction message.
const MaxInvPerTxInvMsg = MaxInvPerMsg

// MsgInvTransaction implements the Message interface and represents a c4ex
// TxInv message. It is used to notify the network about new transactions
// by sending their ID, and let the receiving node decide if it needs it.
type MsgInvTransaction struct {
	baseMessage
	TxIDs []*externalapi.DomainTransactionID
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgInvTransaction) Command() MessageCommand {
	return CmdInvTransaction
}

// NewMsgInvTransaction returns a new c4ex TxInv message that conforms to
// the Message interface. See MsgInvTransaction for details.
func NewMsgInvTransaction(ids []*externalapi.DomainTransactionID) *MsgInvTransaction {
	return &MsgInvTransaction{
		TxIDs: ids,
	}
}
