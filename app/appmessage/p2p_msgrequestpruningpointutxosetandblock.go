package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestPruningPointUTXOSetAndBlock represents a kaspa RequestPruningPointUTXOSetAndBlock message
type MsgRequestPruningPointUTXOSetAndBlock struct {
	baseMessage
	PruningPointHash *externalapi.DomainHash
}

// Command returns the protocol command string for the message
func (msg *MsgRequestPruningPointUTXOSetAndBlock) Command() MessageCommand {
	return CmdRequestPruningPointUTXOSetAndBlock
}

// NewMsgRequestPruningPointUTXOSetAndBlock returns a new MsgRequestPruningPointUTXOSetAndBlock
func NewMsgRequestPruningPointUTXOSetAndBlock(pruningPointHash *externalapi.DomainHash) *MsgRequestPruningPointUTXOSetAndBlock {
	return &MsgRequestPruningPointUTXOSetAndBlock{
		PruningPointHash: pruningPointHash,
	}
}
