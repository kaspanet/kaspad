package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgIBDRootHashMessage implements the Message interface and represents a kaspa
// IBDRootHash message. It is used as a reply to IBD root hash requests.
type MsgIBDRootHashMessage struct {
	baseMessage
	Hash *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgIBDRootHashMessage) Command() MessageCommand {
	return CmdIBDRootHash
}

// NewMsgIBDRootHashMessage returns a new kaspa IBDRootHash message that conforms to
// the Message interface. See MsgIBDRootHashMessage for details.
func NewMsgIBDRootHashMessage(hash *externalapi.DomainHash) *MsgIBDRootHashMessage {
	return &MsgIBDRootHashMessage{
		Hash: hash,
	}
}
