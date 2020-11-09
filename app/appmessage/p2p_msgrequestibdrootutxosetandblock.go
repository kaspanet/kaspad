package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type MsgRequestIBDRootUTXOSetAndBlock struct {
	baseMessage
	IBDRoot *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDRootUTXOSetAndBlock) Command() MessageCommand {
	return CmdRequestIBDRootUTXOSetAndBlock
}

func NewMsgRequestIBDRootUTXOSetAndBlock(ibdRoot *externalapi.DomainHash) *MsgRequestIBDRootUTXOSetAndBlock {
	return &MsgRequestIBDRootUTXOSetAndBlock{
		IBDRoot: ibdRoot,
	}
}
