package appmessage

// MsgRequestNextIBDBlocks implements the Message interface and represents a kaspa
// RequestNextIBDBlocks message. It is used to notify the IBD syncer peer to send
// more IBD blocks.
//
// This message has no payload.
type MsgRequestNextIBDBlocks struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestNextIBDBlocks) Command() MessageCommand {
	return CmdRequestNextIBDBlocks
}

// NewMsgRequestNextIBDBlocks returns a new kaspa RequestNextIBDBlocks message that conforms to the
// Message interface.
func NewMsgRequestNextIBDBlocks() *MsgRequestNextIBDBlocks {
	return &MsgRequestNextIBDBlocks{}
}
