package appmessage

// TxIDBlockPair is an appmessage corresponding to
// its respective RPC message
type TxIDBlockPair struct {
	TxID  string
	Block *RPCBlock
}

// GetAcceptingBlocksOfTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlocksOfTxsRequestMessage struct {
	baseMessage
	TxIDs               []string
	IncludeTransactions bool
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlocksOfTxsRequestMessage) Command() MessageCommand {
	return CmdGetAcceptingBlocksOfTxsRequestMessage
}

// NewGetAcceptingBlocksOfTxsRequest returns a instance of the message
func NewGetAcceptingBlocksOfTxsRequest(txIDs []string, includeTransactions bool) *GetAcceptingBlocksOfTxsRequestMessage {
	return &GetAcceptingBlocksOfTxsRequestMessage{
		TxIDs:               txIDs,
		IncludeTransactions: includeTransactions,
	}
}

// GetAcceptingBlocksOfTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlocksOfTxsResponseMessage struct {
	baseMessage
	TxIDBlockPairs []*TxIDBlockPair

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlocksOfTxsResponseMessage) Command() MessageCommand {
	return CmdGetAcceptingBlocksOfTxsResponseMessage
}

// NewGetAcceptingBlocksOfTxsResponse returns an instance of the message
func NewGetAcceptingBlocksOfTxsResponse(txIDBlockPairs []*TxIDBlockPair) *GetAcceptingBlocksOfTxsResponseMessage {
	return &GetAcceptingBlocksOfTxsResponseMessage{
		TxIDBlockPairs: txIDBlockPairs,
	}
}
