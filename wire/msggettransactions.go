package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MaxInvPerGetTransactionsMsg is the maximum number of hashes that can
// be in a single CmdTxInv message.
const MaxInvPerGetTransactionsMsg = MaxInvPerMsg

// MsgGetRelayBlocks implements the Message interface and represents a kaspa
// GetTransactions message. It is used to request transactions as part of the
// transactions relay protocol.
type MsgGetTransactions struct {
	IDs []*daghash.TxID
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetTransactions) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.IDs)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetTransactions) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.IDs)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetTransactions) Command() MessageCommand {
	return CmdGetTransactions
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetTransactions) MaxPayloadLength(pver uint32) uint32 {
	return daghash.TxIDSize*MaxInvPerGetTransactionsMsg + uint32(VarIntSerializeSize(MaxInvPerGetTransactionsMsg))
}

// NewMsgGetTransactions returns a new kaspa GetTransactions message that conforms to
// the Message interface. See MsgGetTransactions for details.
func NewMsgGetTransactions(ids []*daghash.TxID) *MsgGetTransactions {
	return &MsgGetTransactions{
		IDs: ids,
	}
}
