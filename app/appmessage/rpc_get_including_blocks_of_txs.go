package appmessage

// GetIncludingBlocksOfTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlocksOfTxsRequestMessage struct {
	baseMessage
	TxIDs               []string
	IncludeTransactions bool
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlocksOfTxsRequestMessage) Command() MessageCommand {
	return CmdGetIncludingBlocksOfTxsRequestMessage
}

// NewGetIncludingBlocksOfTxsRequest returns a instance of the message
func NewGetIncludingBlocksOfTxsRequest(txIDs []string, includeTransactions bool) *GetIncludingBlocksOfTxsRequestMessage {
	return &GetIncludingBlocksOfTxsRequestMessage{
		TxIDs:               txIDs,
		IncludeTransactions: includeTransactions,
	}
}

// GetIncludingBlocksOfTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlocksOfTxsResponseMessage struct {
	baseMessage
	TxIDBlockPairs []*TxIDBlockPair

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlocksOfTxsResponseMessage) Command() MessageCommand {
	return CmdGetIncludingBlocksOfTxsResponseMessage
}

// NewGetIncludingBlocksOfTxsResponse returns an instance of the message
func NewGetIncludingBlocksOfTxsResponse(txIDBlockPairs []*TxIDBlockPair) *GetIncludingBlocksOfTxsResponseMessage {
	return &GetIncludingBlocksOfTxsResponseMessage{
		TxIDBlockPairs: txIDBlockPairs,
	}
}
