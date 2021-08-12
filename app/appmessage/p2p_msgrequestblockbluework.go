package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestBlockBlueWork represents a kaspa RequestBlockBlueWork message
type MsgRequestBlockBlueWork struct {
	baseMessage
	Hash *externalapi.DomainHash
}

// Command returns the protocol command string for the message
func (msg *MsgRequestBlockBlueWork) Command() MessageCommand {
	return CmdRequestBlockBlueWork
}

// NewRequestBlockBlueWork returns a new kaspa RequestBlockBlueWork message
func NewRequestBlockBlueWork(hash *externalapi.DomainHash) *MsgRequestBlockBlueWork {
	return &MsgRequestBlockBlueWork{
		Hash: hash,
	}
}
