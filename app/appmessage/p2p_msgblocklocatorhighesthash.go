package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgBlockLocatorHighestHash represents a kaspa BlockLocatorHighestHash message
type MsgBlockLocatorHighestHash struct {
	baseMessage
	HighestHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message
func (msg *MsgBlockLocatorHighestHash) Command() MessageCommand {
	return CmdBlockLocatorHighestHash
}

// NewMsgBlockLocatorHighestHash returns a new BlockLocatorHighestHash message
func NewMsgBlockLocatorHighestHash(highestHash *externalapi.DomainHash) *MsgBlockLocatorHighestHash {
	return &MsgBlockLocatorHighestHash{
		HighestHash: highestHash,
	}
}
