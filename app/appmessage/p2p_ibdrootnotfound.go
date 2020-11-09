package appmessage

// MsgDoneHeaders implements the Message interface and represents a kaspa
// IBDRootNotFound message. It is used to notify the IBD root that was requested
// by other peer was not found.
//
// This message has no payload.
type MsgIBDRootNotFound struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgIBDRootNotFound) Command() MessageCommand {
	return CmdDoneHeaders
}

// NewMsgIBDRootNotFound returns a new kaspa IBDRootNotFound message that conforms to the
// Message interface.
func NewMsgIBDRootNotFound() *MsgDoneHeaders {
	return &MsgDoneHeaders{}
}
