package appmessage

// MsgReject implements the Message interface and represents a kaspa
// Reject message. It is used to notify peers why they are banned.
type MsgReject struct {
	baseMessage
	Reason string
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgReject) Command() MessageCommand {
	return CmdReject
}

// NewMsgReject returns a new kaspa Reject message that conforms to the
// Message interface.
func NewMsgReject(reason string) *MsgReject {
	return &MsgReject{
		Reason: reason,
	}
}
