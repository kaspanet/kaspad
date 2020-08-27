package appmessage

// SubmitBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockRequestMessage struct {
	baseMessage
	BlockHex string
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockRequestMessage) Command() MessageCommand {
	return CmdSubmitBlockRequestMessage
}

// SubmitBlockRequestMessage returns a instance of the message
func NewSubmitBlockRequestMessage(blockHex string) *SubmitBlockRequestMessage {
	return &SubmitBlockRequestMessage{
		BlockHex: blockHex,
	}
}

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
