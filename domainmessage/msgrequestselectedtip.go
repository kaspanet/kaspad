package domainmessage

// MsgRequestSelectedTip implements the Message interface and represents a kaspa
// RequestSelectedTip message. It is used to request the selected tip of another peer.
//
// This message has no payload.
type MsgRequestSelectedTip struct{}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestSelectedTip) Command() MessageCommand {
	return CmdRequestSelectedTip
}

// NewMsgRequestSelectedTip returns a new kaspa RequestSelectedTip message that conforms to the
// Message interface.
func NewMsgRequestSelectedTip() *MsgRequestSelectedTip {
	return &MsgRequestSelectedTip{}
}
