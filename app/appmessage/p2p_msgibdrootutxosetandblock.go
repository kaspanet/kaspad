package appmessage

// MsgIBDRootUTXOSetAndBlock implements the Message interface and represents a kaspa
// IBDRootUTXOSetAndBlock message. It is used to answer RequestIBDRootUTXOSetAndBlock messages.
type MsgIBDRootUTXOSetAndBlock struct {
	baseMessage
	UTXOSet []byte
	Block   *MsgBlock
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgIBDRootUTXOSetAndBlock) Command() MessageCommand {
	return CmdIBDRootUTXOSetAndBlock
}

// NewMsgIBDRootUTXOSetAndBlock returns a new MsgIBDRootUTXOSetAndBlock.
func NewMsgIBDRootUTXOSetAndBlock(utxoSet []byte, block *MsgBlock) *MsgIBDRootUTXOSetAndBlock {
	return &MsgIBDRootUTXOSetAndBlock{
		UTXOSet: utxoSet,
		Block:   block,
	}
}
