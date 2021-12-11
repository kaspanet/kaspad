package appmessage

// GetBalanceByAddressRequest is an appmessage corresponding to
// its respective RPC message
type GetBalanceByAddressRequest struct {
	baseMessage
	Address string
}

// Command returns the protocol command string for the message
func (msg *GetBalanceByAddressRequest) Command() MessageCommand {
	return CmdGetBalanceByAddressRequest
}

// NewGetBalanceByAddressRequest returns a instance of the message
func NewGetBalanceByAddressRequest(address string) *GetBalanceByAddressRequest {
	return &GetBalanceByAddressRequest{
		Address: address,
	}
}

// GetBalanceByAddressResponse is an appmessage corresponding to
// its respective RPC message
type GetBalanceByAddressResponse struct {
	baseMessage
	Balance uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBalanceByAddressResponse) Command() MessageCommand {
	return CmdGetBalanceByAddressResponse
}

// NewGetBalanceByAddressResponse returns a instance of the message
func NewGetBalanceByAddressResponse(Balance uint64) *GetBalanceByAddressResponse {
	return &GetBalanceByAddressResponse{
		Balance: Balance,
	}
}
