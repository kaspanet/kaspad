package appmessage

// GetIncludingBlockOfTxRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlockOfTxRequestMessage struct {
	baseMessage
	TxID                string
	IncludeTransactions bool
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlockOfTxRequestMessage) Command() MessageCommand {
	return CmdGetIncludingBlockOfTxRequestMessage
}

// NewGetIncludingBlockOfTxRequest returns a instance of the message
func NewGetIncludingBlockOfTxRequest(txID string, includeTransactions bool) *GetIncludingBlockOfTxRequestMessage {
	return &GetIncludingBlockOfTxRequestMessage{
		TxID:                txID,
		IncludeTransactions: includeTransactions,
	}
}

// GetIncludingBlockOfTxResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlockOfTxResponseMessage struct {
	baseMessage
	Block *RPCBlock

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlockOfTxResponseMessage) Command() MessageCommand {
	return CmdGetIncludingBlockOfTxResponseMessage
}

// NewGetIncludingBlockOfTxResponse returns an instance of the message
func NewGetIncludingBlockOfTxResponse(block *RPCBlock) *GetIncludingBlockOfTxResponseMessage {
	return &GetIncludingBlockOfTxResponseMessage{
		Block: block,
	}
}
