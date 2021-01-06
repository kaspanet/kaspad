package appmessage

// SubmitBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockRequestMessage struct {
	baseMessage
	Block *MsgBlock
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockRequestMessage) Command() MessageCommand {
	return CmdSubmitBlockRequestMessage
}

// NewSubmitBlockRequestMessage returns a instance of the message
func NewSubmitBlockRequestMessage(block *MsgBlock) *SubmitBlockRequestMessage {
	return &SubmitBlockRequestMessage{
		Block: block,
	}
}

// SubmitBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockResponseMessage) Command() MessageCommand {
	return CmdSubmitBlockResponseMessage
}

// NewSubmitBlockResponseMessage returns a instance of the message
func NewSubmitBlockResponseMessage() *SubmitBlockResponseMessage {
	return &SubmitBlockResponseMessage{}
}
