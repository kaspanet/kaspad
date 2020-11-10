package appmessage

// MsgRequestNextHeaders implements the Message interface and represents a kaspa
// RequestNextHeaders message. It is used to notify the IBD syncer peer to send
// more headers.
//
// This message has no payload.
type MsgRequestNextHeaders struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestNextHeaders) Command() MessageCommand {
	return CmdRequestNextHeaders
}

// NewMsgRequestNextHeaders returns a new kaspa RequestNextHeaders message that conforms to the
// Message interface.
func NewMsgRequestNextHeaders() *MsgRequestNextHeaders {
	return &MsgRequestNextHeaders{}
}
