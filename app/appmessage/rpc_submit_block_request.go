package appmessage

// SubmitBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockRequestMessage) Command() MessageCommand {
	return CmdSubmitBlockRequestMessage
}

// SubmitBlockRequestMessage returns a instance of the message
func NewSubmitBlockRequestMessage() *SubmitBlockRequestMessage {
	return &SubmitBlockRequestMessage{}
}
