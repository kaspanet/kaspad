package appmessage

// MsgRequestIBDRootHash implements the Message interface and represents a kaspa
// MsgRequestIBDRootHash message. It is used to request the IBD root hash
// from a peer during IBD.
//
// This message has no payload.
type MsgRequestIBDRootHash struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestIBDRootHash) Command() MessageCommand {
	return CmdRequestIBDRootHash
}

// NewMsgRequestIBDRootHash returns a new kaspa RequestIBDRootHash message that conforms to the
// Message interface.
func NewMsgRequestIBDRootHash() *MsgRequestIBDRootHash {
	return &MsgRequestIBDRootHash{}
}
