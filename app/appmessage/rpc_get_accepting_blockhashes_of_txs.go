package appmessage



type TxIdBlockHashPair struct {
	TxId string
	blockhash string
}

// GetAcceptingBlockHashesOfTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockHashesOfTxsRequestMessage struct {
	baseMessage
	TxIDs []string
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockHashesOfTxsRequestMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockHashesOfTxsRequestMessage
}

// NewGetAcceptingBlockHashesOfTxsRequest returns a instance of the message
func NewGetAcceptingBlockHashesOfTxsRequest(txIDs []string) *GetAcceptingBlockHashesOfTxsRequestMessage {
	return &GetAcceptingBlockHashesOfTxsRequestMessage{
		TxIDs: txIDs,
	}
}

// GetAcceptingBlockHashesOfTxResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockHashesOfTxsResponseMessage struct {
	baseMessage
	TxIdBlockHashPairs []*TxIdBlockHashPair

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockHashesOfTxsResponseMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockHashesOfTxsResponseMessage
}

// NewGetAcceptingBlockHashesOfTxsResponse returns an instance of the message
func NewGetAcceptingBlockHashesOfTxsResponse(txIdBlockHashPairs []*TxIdBlockHashPair) *GetAcceptingBlockHashesOfTxsResponseMessage {
	return &GetAcceptingBlockHashesOfTxsResponseMessage{
		TxIdBlockHashPairs: txIdBlockHashPairs,
	}
}
