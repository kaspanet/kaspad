package appmessage

// MsgUnexpectedPruningPoint represents a c4ex UnexpectedPruningPoint message
type MsgUnexpectedPruningPoint struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgUnexpectedPruningPoint) Command() MessageCommand {
	return CmdUnexpectedPruningPoint
}

// NewMsgUnexpectedPruningPoint returns a new c4ex UnexpectedPruningPoint message
func NewMsgUnexpectedPruningPoint() *MsgUnexpectedPruningPoint {
	return &MsgUnexpectedPruningPoint{}
}
