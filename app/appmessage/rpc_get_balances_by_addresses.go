package appmessage

// GetBalancesByAddressesRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBalancesByAddressesRequestMessage struct {
	baseMessage
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *GetBalancesByAddressesRequestMessage) Command() MessageCommand {
	return CmdGetBalancesByAddressesRequestMessage
}

// NewGetBalancesByAddressesRequest returns a instance of the message
func NewGetBalancesByAddressesRequest(addresses []string) *GetBalancesByAddressesRequestMessage {
	return &GetBalancesByAddressesRequestMessage{
		Addresses: addresses,
	}
}

// BalancesByAddressesEntry represents the balance of some address
type BalancesByAddressesEntry struct {
	Address string
	Balance uint64
}

// GetBalancesByAddressesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBalancesByAddressesResponseMessage struct {
	baseMessage
	Entries []*BalancesByAddressesEntry

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBalancesByAddressesResponseMessage) Command() MessageCommand {
	return CmdGetBalancesByAddressesResponseMessage
}

// NewGetBalancesByAddressesResponse returns an instance of the message
func NewGetBalancesByAddressesResponse(entries []*BalancesByAddressesEntry) *GetBalancesByAddressesResponseMessage {
	return &GetBalancesByAddressesResponseMessage{
		Entries: entries,
	}
}
