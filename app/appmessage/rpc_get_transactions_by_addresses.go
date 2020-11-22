package appmessage

// GetTransactionsByAddressesRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetTransactionsByAddressesRequestMessage struct {
	baseMessage
	StartingBlockHash string
	Addresses         []string
}

// Command returns the protocol command string for the message
func (msg *GetTransactionsByAddressesRequestMessage) Command() MessageCommand {
	return CmdGetTransactionsByAddressesRequestMessage
}

// NewGetTransactionsByAddressesRequestMessage returns a instance of the message
func NewGetTransactionsByAddressesRequestMessage(startingBlockHash string, addresses []string) *GetTransactionsByAddressesRequestMessage {
	return &GetTransactionsByAddressesRequestMessage{
		StartingBlockHash: startingBlockHash,
		Addresses:         addresses,
	}
}

// GetTransactionsByAddressesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetTransactionsByAddressesResponseMessage struct {
	baseMessage
	LasBlockScanned string
	Transactions    []*TransactionVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetTransactionsByAddressesResponseMessage) Command() MessageCommand {
	return CmdGetTransactionsByAddressesResponseMessage
}

// NewGetTransactionsByAddressesResponseMessage returns a instance of the message
func NewGetTransactionsByAddressesResponseMessage(lasBlockScanned string, transactions []*TransactionVerboseData) *GetTransactionsByAddressesResponseMessage {
	return &GetTransactionsByAddressesResponseMessage{
		LasBlockScanned: lasBlockScanned,
		Transactions:    transactions,
	}
}
