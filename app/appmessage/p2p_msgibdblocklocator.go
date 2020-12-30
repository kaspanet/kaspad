package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgIBDBlockLocator represents a kaspa ibdBlockLocator message
type MsgIBDBlockLocator struct {
	baseMessage
	Hashes []*externalapi.DomainHash
}

// Command returns the protocol command string for the message
func (msg *MsgIBDBlockLocator) Command() MessageCommand {
	return CmdIBDBlockLocator
}

// NewMsgIBDBlockLocator returns a new kaspa ibdBlockLocator message
func NewMsgIBDBlockLocator(hashes []*externalapi.DomainHash) *MsgIBDBlockLocator {
	return &MsgIBDBlockLocator{
		Hashes: hashes,
	}
}
