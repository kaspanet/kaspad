package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MsgGetRelayBlock implements the Message interface and represents a kaspa
// block inventory message. It is used to request a block with a given hash.
type MsgGetRelayBlock struct {
	Hash *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetRelayBlock) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.Hash)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetRelayBlock) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.Hash)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetRelayBlock) Command() string {
	return CmdGetRelayBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetRelayBlock) MaxPayloadLength(pver uint32) uint32 {
	return daghash.HashSize
}

// NewMsgInvBlock returns a new kaspa invrelblk message that conforms to
// the Message interface. See MsgGetRelayBlock for details.
func NewMsgGetRelayBlock(hash *daghash.Hash) *MsgGetRelayBlock {
	return &MsgGetRelayBlock{
		Hash: hash,
	}
}
