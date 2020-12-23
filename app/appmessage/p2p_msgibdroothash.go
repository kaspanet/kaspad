package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgIBDRootHash implements the Message interface and represents a kaspa
// IBDRootHash message. It is used as a reply to IBD root hash requests.
type MsgIBDRootHash struct {
	baseMessage
	Hash *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgIBDRootHash) Command() MessageCommand {
	return CmdIBDRootHash
}

// NewMsgIBDRootHash returns a new kaspa IBDRootHash message that conforms to
// the Message interface. See MsgIBDRootHash for details.
func NewMsgIBDRootHash(hash *externalapi.DomainHash) *MsgIBDRootHash {
	return &MsgIBDRootHash{
		Hash: hash,
	}
}
