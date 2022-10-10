package appmessage

// GetAcceptingBlockOfTxRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockOfTxRequestMessage struct {
	baseMessage
	TxID string
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockOfTxRequestMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockHashOfTxRequestMessage
}

// NewGetAcceptingBlockOfTxRequest returns a instance of the message
func NewGetAcceptingBlockOfTxRequest(txID string) *GetAcceptingBlockOfTxRequestMessage {
	return &GetAcceptingBlockOfTxRequestMessage{
		TxID: txID,
	}
}

// GetAcceptingBlockOfTxResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockOfTxResponseMessage struct {
	baseMessage
	Block *RPCBlock

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockOfTxResponseMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockOfTxResponseMessage
}

// NewGetAcceptingBlockOfTxResponse returns an instance of the message
func NewGetAcceptingBlockOfTxResponse(block *RPCBlock) *GetAcceptingBlockOfTxResponseMessage {
	return &GetAcceptingBlockOfTxResponseMessage{
		Block: block,
	}
}
