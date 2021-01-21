package appmessage

// MsgRequestPruningPointHashMessage represents a kaspa RequestPruningPointHashMessage message
type MsgRequestPruningPointHashMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgRequestPruningPointHashMessage) Command() MessageCommand {
	return CmdRequestPruningPointHash
}

// NewMsgRequestPruningPointHashMessage returns a new kaspa RequestPruningPointHash message
func NewMsgRequestPruningPointHashMessage() *MsgRequestPruningPointHashMessage {
	return &MsgRequestPruningPointHashMessage{}
}
