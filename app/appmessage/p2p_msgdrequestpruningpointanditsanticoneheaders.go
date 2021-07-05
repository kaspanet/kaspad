package appmessage

// MsgRequestPruningPointAndItsAnticone represents a kaspa RequestPruningPointAndItsAnticone message
type MsgRequestPruningPointAndItsAnticone struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgRequestPruningPointAndItsAnticone) Command() MessageCommand {
	return CmdRequestPruningPointAndItsAnticone
}

// NewMsgRequestPruningPointAndItsAnticone returns a new MsgRequestPruningPointAndItsAnticone.
func NewMsgRequestPruningPointAndItsAnticone() *MsgRequestPruningPointAndItsAnticone {
	return &MsgRequestPruningPointAndItsAnticone{}
}
