package appmessage

// StopRequestMessage is an appmessage corresponding to
// its respective RPC message
type StopRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *StopRequestMessage) Command() MessageCommand {
	return CmdStopRequestMessage
}

// NewStopRequestMessage returns a instance of the message
func NewStopRequestMessage() *StopRequestMessage {
	return &StopRequestMessage{}
}

// StopResponseMessage is an appmessage corresponding to
// its respective RPC message
type StopResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *StopResponseMessage) Command() MessageCommand {
	return CmdStopResponseMessage
}

// NewStopResponseMessage returns a instance of the message
func NewStopResponseMessage() *StopResponseMessage {
	return &StopResponseMessage{}
}
