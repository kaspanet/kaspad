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

// NewGetBalanceByAddressRequestMessage returns a instance of the message
func NewGetBalanceByAddressRequestMessage(address string) *GetBalanceByAddressRequestMessage {
	return &GetBalanceByAddressRequestMessage{
		Address: address,
	}
}

// GetBalanceByAddressResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBalanceByAddressResponseMessage struct {
	baseMessage
	balance uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBalanceByAddressResponseMessage) Command() MessageCommand {
	return CmdGetBalanceByAddressResponseMessage
}

// NewGetBalanceByAddressResponseMessage returns a instance of the message
func NewGetBalanceByAddressResponseMessage(balance uint64) *GetBalanceByAddressResponseMessage {
	return &GetBalanceByAddressResponseMessage{
		balance: balance,
	}
}
