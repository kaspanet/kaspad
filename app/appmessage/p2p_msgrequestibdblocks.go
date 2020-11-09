package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

const MaxRequestIBDBlocksHashes = MaxInvPerMsg

type MsgRequestIBDBlocks struct {
	baseMessage
	Hashes []*externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDBlocks) Command() MessageCommand {
	return CmdRequestIBDBlocks
}

func NewMsgRequestIBDBlocks(hashes []*externalapi.DomainHash) *MsgRequestIBDBlocks {
	return &MsgRequestIBDBlocks{
		Hashes: hashes,
	}
}
