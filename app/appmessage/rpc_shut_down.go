package appmessage

// ShutDownRequestMessage is an appmessage corresponding to
// its respective RPC message
type ShutDownRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *ShutDownRequestMessage) Command() MessageCommand {
	return CmdShutDownRequestMessage
}

// NewShutDownRequestMessage returns a instance of the message
func NewShutDownRequestMessage() *ShutDownRequestMessage {
	return &ShutDownRequestMessage{}
}

// ShutDownResponseMessage is an appmessage corresponding to
// its respective RPC message
type ShutDownResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *ShutDownResponseMessage) Command() MessageCommand {
	return CmdShutDownResponseMessage
}

// NewShutDownResponseMessage returns a instance of the message
func NewShutDownResponseMessage() *ShutDownResponseMessage {
	return &ShutDownResponseMessage{}
}
