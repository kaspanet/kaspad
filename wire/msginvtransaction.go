package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MaxInvPerTxInvMsg is the maximum number of hashes that can
// be in a single CmdInvTransaction message.
const MaxInvPerTxInvMsg = MaxInvPerMsg

// MsgInvTransaction implements the Message interface and represents a kaspa
// TxInv message. It is used to notify the network about new transactions
// by sending their ID, and let the receiving node decide if it needs it.
type MsgInvTransaction struct {
	TXIDs []*daghash.TxID
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgInvTransaction) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.TXIDs)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgInvTransaction) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.TXIDs)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgInvTransaction) Command() MessageCommand {
	return CmdInvTransaction
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgInvTransaction) MaxPayloadLength(pver uint32) uint32 {
	return daghash.TxIDSize*MaxInvPerTxInvMsg + uint32(VarIntSerializeSize(MaxInvPerTxInvMsg))
}

// NewMsgTxInv returns a new kaspa TxInv message that conforms to
// the Message interface. See MsgInvTransaction for details.
func NewMsgTxInv(ids []*daghash.TxID) *MsgInvTransaction {
	return &MsgInvTransaction{
		TXIDs: ids,
	}
}
