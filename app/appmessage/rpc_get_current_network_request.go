package appmessage

// GetCurrentNetworkRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetCurrentNetworkRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetCurrentNetworkRequestMessage) Command() MessageCommand {
	return CmdGetCurrentNetworkRequestMessage
}

// GetCurrentNetworkRequestMessage returns a instance of the message
func NewGetCurrentVersionRequestMessage() *GetCurrentNetworkRequestMessage {
	return &GetCurrentNetworkRequestMessage{}
}
