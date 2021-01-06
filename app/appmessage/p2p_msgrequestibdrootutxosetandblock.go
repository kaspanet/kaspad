package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgRequestIBDRootUTXOSetAndBlock implements the Message interface and represents a kaspa
// RequestIBDRootUTXOSetAndBlock message. It is used to request the UTXO set and block body
// of the IBD root block.
type MsgRequestIBDRootUTXOSetAndBlock struct {
	baseMessage
	IBDRoot *externalapi.DomainHash
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDRootUTXOSetAndBlock) Command() MessageCommand {
	return CmdRequestIBDRootUTXOSetAndBlock
}

// NewMsgRequestIBDRootUTXOSetAndBlock returns a new MsgRequestIBDRootUTXOSetAndBlock.
func NewMsgRequestIBDRootUTXOSetAndBlock(ibdRoot *externalapi.DomainHash) *MsgRequestIBDRootUTXOSetAndBlock {
	return &MsgRequestIBDRootUTXOSetAndBlock{
		IBDRoot: ibdRoot,
	}
}
