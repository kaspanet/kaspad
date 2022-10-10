package appmessage

// GetTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetTxsRequestMessage struct {
	baseMessage
	TxIDs []string
}

// Command returns the protocol command string for the message
func (msg *GetTxsRequestMessage) Command() MessageCommand {
	return CmdGetTxsRequestMessage
}

// NewGetTxsRequest returns a instance of the message
func NewGetTxsRequest(txIDs []string) *GetTxsRequestMessage {
	return &GetTxsRequestMessage{
		TxIDs: txIDs,
	}
}

// GetTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetTxsResponseMessage struct {
	baseMessage
	Transactions []*RPCTransaction

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetTxsResponseMessage) Command() MessageCommand {
	return CmdGetTxsResponseMessage
}

// NewGetTxsResponse returns an instance of the message
func NewGetTxsResponse(transactions []*RPCTransaction) *GetTxsResponseMessage {
	return &GetTxsResponseMessage{
		Transactions: transactions,
	}
}
