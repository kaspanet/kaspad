package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MsgInvRelayBlock implements the Message interface and represents a kaspa
// block inventory message. It is used to notify the network about new block
// by sending their hash, and let the receiving node decide if it needs it.
type MsgInvRelayBlock struct {
	Hash *daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgInvRelayBlock) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.Hash)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgInvRelayBlock) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.Hash)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgInvRelayBlock) Command() string {
	return CmdInvRelayBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgInvRelayBlock) MaxPayloadLength(pver uint32) uint32 {
	return daghash.HashSize
}

// NewMsgInvBlock returns a new kaspa invrelblk message that conforms to
// the Message interface. See MsgInvRelayBlock for details.
func NewMsgInvBlock(hash *daghash.Hash) *MsgInvRelayBlock {
	return &MsgInvRelayBlock{
		Hash: hash,
	}
}
