package domainmessage

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgRequestRelayBlocksHashes is the maximum number of hashes that can
// be in a single RequestRelayBlocks message.
const MsgRequestRelayBlocksHashes = MaxInvPerMsg

// MsgRequestRelayBlocks implements the Message interface and represents a kaspa
// RequestRelayBlocks message. It is used to request blocks as part of the block
// relay protocol.
type MsgRequestRelayBlocks struct {
	baseMessage
	Hashes []*daghash.Hash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestRelayBlocks) Command() MessageCommand {
	return CmdRequestRelayBlocks
}

// NewMsgRequestRelayBlocks returns a new kaspa RequestRelayBlocks message that conforms to
// the Message interface. See MsgRequestRelayBlocks for details.
func NewMsgRequestRelayBlocks(hashes []*daghash.Hash) *MsgRequestRelayBlocks {
	return &MsgRequestRelayBlocks{
		Hashes: hashes,
	}
}
