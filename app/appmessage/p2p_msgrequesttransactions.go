package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MaxInvPerRequestTransactionsMsg is the maximum number of hashes that can
// be in a single CmdInvTransaction message.
const MaxInvPerRequestTransactionsMsg = MaxInvPerMsg

// MsgRequestTransactions implements the Message interface and represents a kaspa
// RequestTransactions message. It is used to request transactions as part of the
// transactions relay protocol.
type MsgRequestTransactions struct {
	baseMessage
	IDs []*externalapi.DomainTransactionID
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestTransactions) Command() MessageCommand {
	return CmdRequestTransactions
}

// NewMsgRequestTransactions returns a new kaspa RequestTransactions message that conforms to
// the Message interface. See MsgRequestTransactions for details.
func NewMsgRequestTransactions(ids []*externalapi.DomainTransactionID) *MsgRequestTransactions {
	return &MsgRequestTransactions{
		IDs: ids,
	}
}
