package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestIBDChainBlockLocator implements the Message interface and represents a kaspa
// IBDRequestChainBlockLocator message. It is used to request a block locator between low
// and high hash.
// The locator is returned via a locator message (MsgIBDChainBlockLocator).
type MsgRequestIBDChainBlockLocator struct {
	baseMessage
	HighHash *externalapi.DomainHash
	LowHash  *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDChainBlockLocator) Command() MessageCommand {
	return CmdRequestIBDChainBlockLocator
}

// NewMsgIBDRequestChainBlockLocator returns a new IBDRequestChainBlockLocator message that conforms to the
// Message interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgIBDRequestChainBlockLocator(highHash, lowHash *externalapi.DomainHash) *MsgRequestIBDChainBlockLocator {
	return &MsgRequestIBDChainBlockLocator{
		HighHash: highHash,
		LowHash:  lowHash,
	}
}
