package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
	"io"
)

// MsgGetRelayBlocksHashes is the maximum number of hashes that can
// be in a single getrelblks message.
const MsgGetRelayBlocksHashes = MaxInvPerMsg

// MsgGetRelayBlocks implements the Message interface and represents a kaspa
// getrelblks message. It is used to request blocks as part of the block
// relay protocol.
type MsgGetRelayBlocks struct {
	Hashes []*daghash.Hash
}

// KaspaDecode decodes r using the kaspa protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgGetRelayBlocks) KaspaDecode(r io.Reader, pver uint32) error {
	return ReadElement(r, &msg.Hashes)
}

// KaspaEncode encodes the receiver to w using the kaspa protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgGetRelayBlocks) KaspaEncode(w io.Writer, pver uint32) error {
	return WriteElement(w, msg.Hashes)
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetRelayBlocks) Command() string {
	return CmdGetRelayBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver. This is part of the Message interface implementation.
func (msg *MsgGetRelayBlocks) MaxPayloadLength(pver uint32) uint32 {
	return daghash.HashSize
}

// NewMsgGetRelayBlocks returns a new kaspa getrelblks message that conforms to
// the Message interface. See MsgGetRelayBlocks for details.
func NewMsgGetRelayBlocks(hashes []*daghash.Hash) *MsgGetRelayBlocks {
	return &MsgGetRelayBlocks{
		Hashes: hashes,
	}
}
