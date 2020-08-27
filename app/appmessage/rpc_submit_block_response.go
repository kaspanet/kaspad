package appmessage

// SubmitBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockResponseMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockResponseMessage) Command() MessageCommand {
	return CmdSubmitBlockResponseMessage
}

// SubmitBlockResponseMessage returns a instance of the message
func NewSubmitBlockResponseMessage() *SubmitBlockResponseMessage {
	return &SubmitBlockResponseMessage{}
}
