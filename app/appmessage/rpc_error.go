package appmessage

// RPCErrorMessage is an appmessage corresponding to
// its respective RPC message
type RPCErrorMessage struct {
	baseMessage
	Message string
}

// Command returns the protocol command string for the message
func (msg *RPCErrorMessage) Command() MessageCommand {
	return CmdRPCErrorMessage
}

// RPCErrorMessage returns a instance of the message
func NewRPCErrorMessage(message string) *RPCErrorMessage {
	return &RPCErrorMessage{
		Message: message,
	}
}
