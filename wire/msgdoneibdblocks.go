package wire

// MsgDoneIBDBlocks implements the Message interface and represents a kaspa
// DoneIBDBlocks message. It is used to notify the IBD syncing peer that the
// syncer sent all the requested blocks.
//
// This message has no payload.
type MsgDoneIBDBlocks struct{}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgDoneIBDBlocks) Command() MessageCommand {
	return CmdDoneIBDBlocks
}

// NewMsgDoneIBDBlocks returns a new kaspa GetNextIBDBlocks message that conforms to the
// Message interface.
func NewMsgDoneIBDBlocks() *MsgDoneIBDBlocks {
	return &MsgDoneIBDBlocks{}
}
