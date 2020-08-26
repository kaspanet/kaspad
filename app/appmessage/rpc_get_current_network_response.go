package appmessage

// GetCurrentVersionResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetCurrentVersionResponseMessage struct {
	baseMessage
	currentNetwork string
}

// Command returns the protocol command string for the message
func (msg *GetCurrentVersionResponseMessage) Command() MessageCommand {
	return CmdGetCurrentVersionResponseMessage
}

// GetCurrentVersionResponseMessage returns a instance of the message
func NewGetCurrentVersionResponseMessage(currentNetwork string) *GetCurrentVersionResponseMessage {
	return &GetCurrentVersionResponseMessage{
		currentNetwork: currentNetwork,
	}
}
