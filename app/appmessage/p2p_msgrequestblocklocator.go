package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestBlockLocator implements the Message interface and represents a kaspa
// RequestBlockLocator message. It is used to request a block locator between high
// and low hash.
// The locator is returned via a locator message (MsgBlockLocator).
type MsgRequestBlockLocator struct {
	baseMessage
	HighHash *externalapi.DomainHash
	LowHash  *externalapi.DomainHash
	Limit    uint32
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestBlockLocator) Command() MessageCommand {
	return CmdRequestBlockLocator
}

// NewMsgRequestBlockLocator returns a new RequestBlockLocator message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgRequestBlockLocator(highHash, lowHash *externalapi.DomainHash, limit uint32) *MsgRequestBlockLocator {
	return &MsgRequestBlockLocator{
		HighHash: highHash,
		LowHash:  lowHash,
		Limit:    limit,
	}
}
