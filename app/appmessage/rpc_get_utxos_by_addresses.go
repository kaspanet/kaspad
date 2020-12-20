package appmessage

// GetUTXOsByAddressesRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetUTXOsByAddressesRequestMessage struct {
	baseMessage
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *GetUTXOsByAddressesRequestMessage) Command() MessageCommand {
	return CmdGetUTXOsByAddressesRequestMessage
}

// NewGetUTXOsByAddressesRequestMessage returns a instance of the message
func NewGetUTXOsByAddressesRequestMessage(addresses []string) *GetUTXOsByAddressesRequestMessage {
	return &GetUTXOsByAddressesRequestMessage{
		Addresses: addresses,
	}
}

// GetUTXOsByAddressesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetUTXOsByAddressesResponseMessage struct {
	baseMessage
	Entries []*UTXOsByAddressesEntry

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetUTXOsByAddressesResponseMessage) Command() MessageCommand {
	return CmdGetUTXOsByAddressesResponseMessage
}

// NewGetUTXOsByAddressesResponseMessage returns a instance of the message
func NewGetUTXOsByAddressesResponseMessage(entries []*UTXOsByAddressesEntry) *GetUTXOsByAddressesResponseMessage {
	return &GetUTXOsByAddressesResponseMessage{
		Entries: entries,
	}
}
