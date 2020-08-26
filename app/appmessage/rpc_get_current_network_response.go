package appmessage

// GetCurrentNetworkResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetCurrentNetworkResponseMessage struct {
	baseMessage
	CurrentNetwork string
}

// Command returns the protocol command string for the message
func (msg *GetCurrentNetworkResponseMessage) Command() MessageCommand {
	return CmdGetCurrentNetworkResponseMessage
}

// GetCurrentNetworkResponseMessage returns a instance of the message
func NewGetCurrentVersionResponseMessage(currentNetwork string) *GetCurrentNetworkResponseMessage {
	return &GetCurrentNetworkResponseMessage{
		CurrentNetwork: currentNetwork,
	}
}
