package appmessage

// MsgDoneBlocksWithMetaData implements the Message interface and represents a kaspa
// DoneBlocksWithMetaData message
//
// This message has no payload.
type MsgDoneBlocksWithMetaData struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgDoneBlocksWithMetaData) Command() MessageCommand {
	return CmdDoneBlocksWithMetaData
}

// NewMsgDoneBlocksWithMetaData returns a new kaspa DoneBlocksWithMetaData message that conforms to the
// Message interface.
func NewMsgDoneBlocksWithMetaData() *MsgDoneBlocksWithMetaData {
	return &MsgDoneBlocksWithMetaData{}
}
