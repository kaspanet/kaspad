package appmessage

// TxIDBlockHashPair is an appmessage corresponding to
// its respective RPC message
type TxIDBlockHashPair struct {
	TxID      string
	Hash 	  string
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

// GetAcceptingBlockHashesOfTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockHashesOfTxsResponseMessage struct {
	baseMessage
	TxIDBlockHashPairs []*TxIDBlockHashPair

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockHashesOfTxsResponseMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockHashesOfTxsResponseMessage
}

// NewGetAcceptingBlockHashesOfTxsResponse returns an instance of the message
func NewGetAcceptingBlockHashesOfTxsResponse(txIDBlockHashPairs []*TxIDBlockHashPair) *GetAcceptingBlockHashesOfTxsResponseMessage {
	return &GetAcceptingBlockHashesOfTxsResponseMessage{
		TxIDBlockHashPairs: txIDBlockHashPairs,
	}
}
