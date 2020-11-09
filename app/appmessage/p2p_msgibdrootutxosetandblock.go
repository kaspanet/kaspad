package appmessage

type MsgIBDRootUTXOSetAndBlock struct {
	baseMessage
	UTXOSet []byte
	Block   *MsgBlock
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgIBDRootUTXOSetAndBlock) Command() MessageCommand {
	return CmdRequestIBDRootUTXOSetAndBlock
}

func NewMsgIBDRootUTXOSetAndBlock(utxoSet []byte, block *MsgBlock) *MsgIBDRootUTXOSetAndBlock {
	return &MsgIBDRootUTXOSetAndBlock{
		UTXOSet: utxoSet,
		Block:   block,
	}
}
