package appmessage

// GetBalanceByAddressRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBalanceByAddressRequestMessage struct {
	baseMessage
	Address string
}

// Command returns the protocol command string for the message
func (msg *GetBalanceByAddressRequestMessage) Command() MessageCommand {
	return CmdGetBalanceByAddressRequestMessage
}

// NewGetBalanceByAddressRequest returns a instance of the message
func NewGetBalanceByAddressRequest(address string) *GetBalanceByAddressRequestMessage {
	return &GetBalanceByAddressRequestMessage{
		Address: address,
	}
}

// GetBalanceByAddressResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBalanceByAddressResponseMessage struct {
	baseMessage
	Balance uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBalanceByAddressResponseMessage) Command() MessageCommand {
	return CmdGetBalanceByAddressResponseMessage
}

// NewGetBalanceByAddressResponse returns an instance of the message
func NewGetBalanceByAddressResponse(Balance uint64) *GetBalanceByAddressResponseMessage {
	return &GetBalanceByAddressResponseMessage{
		Balance: Balance,
	}
}
