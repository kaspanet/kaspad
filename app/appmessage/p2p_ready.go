package appmessage

// MsgReady implements the Message interface and represents a kaspa
// Ready message. It is used to notify that the peer is ready to receive
// messages.
//
// This message has no payload.
type MsgReady struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgReady) Command() MessageCommand {
	return CmdReady
}

// NewMsgReady returns a new kaspa Ready message that conforms to the
// Message interface.
func NewMsgReady() *MsgReady {
	return &MsgReady{}
}
