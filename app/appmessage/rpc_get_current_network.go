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

// NewGetCurrentNetworkRequestMessage returns a instance of the message
func NewGetCurrentNetworkRequestMessage() *GetCurrentNetworkRequestMessage {
	return &GetCurrentNetworkRequestMessage{}
}

// GetCurrentNetworkResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetCurrentNetworkResponseMessage struct {
	baseMessage
	CurrentNetwork string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetCurrentNetworkResponseMessage) Command() MessageCommand {
	return CmdGetCurrentNetworkResponseMessage
}

// NewGetCurrentNetworkResponseMessage returns a instance of the message
func NewGetCurrentNetworkResponseMessage(currentNetwork string) *GetCurrentNetworkResponseMessage {
	return &GetCurrentNetworkResponseMessage{
		CurrentNetwork: currentNetwork,
	}
}
