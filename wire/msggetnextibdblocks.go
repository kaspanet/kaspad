package wire

// MsgGetNextIBDBlocks implements the Message interface and represents a kaspa
// GetNextIBDBlocks message. It is used to notify the IBD syncer peer to send
// more blocks.
//
// This message has no payload.
type MsgGetNextIBDBlocks struct{}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgGetNextIBDBlocks) Command() MessageCommand {
	return CmdGetNextIBDBlocks
}

// NewMsgGetNextIBDBlocks returns a new kaspa GetNextIBDBlocks message that conforms to the
// Message interface.
func NewMsgGetNextIBDBlocks() *MsgGetNextIBDBlocks {
	return &MsgGetNextIBDBlocks{}
}
