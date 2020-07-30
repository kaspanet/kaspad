package wire

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgGetRelayBlocksHashes is the maximum number of hashes that can
// be in a single RequestRelayBlocks message.
const MsgGetRelayBlocksHashes = MaxInvPerMsg

// MsgRequestRelayBlocks implements the Message interface and represents a kaspa
// RequestRelayBlocks message. It is used to request blocks as part of the block
// relay protocol.
type MsgRequestRelayBlocks struct {
	Hashes []*daghash.Hash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestRelayBlocks) Command() MessageCommand {
	return CmdRequestRelayBlocks
}

// NewMsgGetRelayBlocks returns a new kaspa RequestRelayBlocks message that conforms to
// the Message interface. See MsgRequestRelayBlocks for details.
func NewMsgGetRelayBlocks(hashes []*daghash.Hash) *MsgRequestRelayBlocks {
	return &MsgRequestRelayBlocks{
		Hashes: hashes,
	}
}
