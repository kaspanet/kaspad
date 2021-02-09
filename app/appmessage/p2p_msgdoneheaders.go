package appmessage

// MsgDoneHeaders implements the Message interface and represents a kaspa
// DoneHeaders message. It is used to notify the IBD syncing peer that the
// syncer sent all the requested headers.
//
// This message has no payload.
type MsgDoneHeaders struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgDoneHeaders) Command() MessageCommand {
	return CmdDoneHeaders
}

// NewMsgDoneHeaders returns a new kaspa DoneIBDBlocks message that conforms to the
// Message interface.
func NewMsgDoneHeaders() *MsgDoneHeaders {
	return &MsgDoneHeaders{}
}
