package appmessage

// GetIncludingBlockHashesOfTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlockHashesOfTxsRequestMessage struct {
	baseMessage
	TxIDs []string
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlockHashesOfTxsRequestMessage) Command() MessageCommand {
	return CmdGetIncludingBlockHashesOfTxsRequestMessage
}

// NewGetIncludingBlockHashesOfTxsRequest returns a instance of the message
func NewGetIncludingBlockHashesOfTxsRequest(txIDs []string) *GetIncludingBlockHashesOfTxsRequestMessage {
	return &GetIncludingBlockHashesOfTxsRequestMessage{
		TxIDs: txIDs,
	}
}

// GetIncludingBlockHashesOfTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlockHashesOfTxsResponseMessage struct {
	baseMessage
	TxIDBlockHashPairs []*TxIDBlockHashPair

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlockHashesOfTxsResponseMessage) Command() MessageCommand {
	return CmdGetIncludingBlockHashesOfTxsResponseMessage
}

// NewGetIncludingBlockHashesOfTxsResponse returns an instance of the message
func NewGetIncludingBlockHashesOfTxsResponse(txIDBlockHashPairs []*TxIDBlockHashPair) *GetIncludingBlockHashesOfTxsResponseMessage {
	return &GetIncludingBlockHashesOfTxsResponseMessage{
		TxIDBlockHashPairs: txIDBlockHashPairs,
	}
}
