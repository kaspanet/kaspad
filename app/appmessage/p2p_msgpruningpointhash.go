package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgPruningPointHashMessage represents a kaspa PruningPointHash message
type MsgPruningPointHashMessage struct {
	baseMessage
	Hash *externalapi.DomainHash
}

// Command returns the protocol command string for the message
func (msg *MsgPruningPointHashMessage) Command() MessageCommand {
	return CmdPruningPointHash
}

// NewPruningPointHashMessage returns a new kaspa PruningPointHash message
func NewPruningPointHashMessage(hash *externalapi.DomainHash) *MsgPruningPointHashMessage {
	return &MsgPruningPointHashMessage{
		Hash: hash,
	}
}
