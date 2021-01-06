package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgIBDBlockLocatorHighestHash represents a kaspa BlockLocatorHighestHash message
type MsgIBDBlockLocatorHighestHash struct {
	baseMessage
	HighestHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message
func (msg *MsgIBDBlockLocatorHighestHash) Command() MessageCommand {
	return CmdIBDBlockLocatorHighestHash
}

// NewMsgIBDBlockLocatorHighestHash returns a new BlockLocatorHighestHash message
func NewMsgIBDBlockLocatorHighestHash(highestHash *externalapi.DomainHash) *MsgIBDBlockLocatorHighestHash {
	return &MsgIBDBlockLocatorHighestHash{
		HighestHash: highestHash,
	}
}
