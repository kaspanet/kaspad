package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MaxInvPerGetTransactionsMsg is the maximum number of hashes that can
// be in a single CmdInvTransaction message.
const MaxInvPerGetTransactionsMsg = MaxInvPerMsg

// MsgRequestTransactions implements the Message interface and represents a kaspa
// RequestTransactions message. It is used to request transactions as part of the
// transactions relay protocol.
type MsgRequestTransactions struct {
	IDs []*daghash.TxID
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgRequestTransactions) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.IDs)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgRequestTransactions) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.IDs)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestTransactions) Command() MessageCommand {
	return CmdRequestTransactions
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgRequestTransactions) MaxPayloadLength(pver uint32) uint32 {
	return daghash.TxIDSize*MaxInvPerGetTransactionsMsg + uint32(VarIntSerializeSize(MaxInvPerGetTransactionsMsg))
}

// NewMsgGetTransactions returns a new kaspa RequestTransactions message that conforms to
// the Message interface. See MsgRequestTransactions for details.
func NewMsgGetTransactions(ids []*daghash.TxID) *MsgRequestTransactions {
	return &MsgRequestTransactions{
		IDs: ids,
	}
}
