package appmessage

// MsgRequestIBDRootHashMessage implements the Message interface and represents a kaspa
// MsgRequestIBDRootHashMessage message. It is used to request the IBD root hash
// from a peer during IBD.
//
// This message has no payload.
type MsgRequestIBDRootHashMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDRootHashMessage) Command() MessageCommand {
	return CmdRequestIBDRootHash
}

// NewMsgRequestIBDRootHashMessage returns a new kaspa RequestIBDRootHash message that conforms to the
// Message interface.
func NewMsgRequestIBDRootHashMessage() *MsgRequestIBDRootHashMessage {
	return &MsgRequestIBDRootHashMessage{}
}
