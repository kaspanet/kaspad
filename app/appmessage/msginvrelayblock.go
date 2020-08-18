package appmessage

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// MsgInvRelayBlock implements the Message interface and represents a kaspa
// block inventory message. It is used to notify the network about new block
// by sending their hash, and let the receiving node decide if it needs it.
type MsgInvRelayBlock struct {
	baseMessage
	Hash *daghash.Hash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgInvRelayBlock) Command() MessageCommand {
	return CmdInvRelayBlock
}

// NewMsgInvBlock returns a new kaspa invrelblk message that conforms to
// the Message interface. See MsgInvRelayBlock for details.
func NewMsgInvBlock(hash *daghash.Hash) *MsgInvRelayBlock {
	return &MsgInvRelayBlock{
		Hash: hash,
	}
}
