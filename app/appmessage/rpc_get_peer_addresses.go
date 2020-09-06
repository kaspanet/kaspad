package appmessage

// GetPeerAddressesRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetPeerAddressesRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetPeerAddressesRequestMessage) Command() MessageCommand {
	return CmdGetPeerAddressesRequestMessage
}

// NewGetPeerAddressesRequestMessage returns a instance of the message
func NewGetPeerAddressesRequestMessage() *GetPeerAddressesRequestMessage {
	return &GetPeerAddressesRequestMessage{}
}

// GetPeerAddressesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetPeerAddressesResponseMessage struct {
	baseMessage
	Addresses []*GetPeerAddressesKnownAddressMessage

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetPeerAddressesResponseMessage) Command() MessageCommand {
	return CmdGetPeerAddressesResponseMessage
}

// NewGetPeerAddressesResponseMessage returns a instance of the message
func NewGetPeerAddressesResponseMessage(addresses []*GetPeerAddressesKnownAddressMessage) *GetPeerAddressesResponseMessage {
	return &GetPeerAddressesResponseMessage{
		Addresses: addresses,
	}
}

// GetPeerAddressesKnownAddressMessage is an appmessage corresponding to
// its respective RPC message
type GetPeerAddressesKnownAddressMessage struct {
	Addr string
}
