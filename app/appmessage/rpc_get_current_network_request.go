package appmessage

// GetCurrentVersionRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetCurrentVersionRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetCurrentVersionRequestMessage) Command() MessageCommand {
	return CmdGetCurrentVersionRequestMessage
}

// GetCurrentVersionRequestMessage returns a instance of the message
func NewGetCurrentVersionRequestMessage() *GetCurrentVersionRequestMessage {
	return &GetCurrentVersionRequestMessage{}
}
