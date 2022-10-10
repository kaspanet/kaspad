package appmessage

// GetIncludingBlockHashOfTxRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlockHashOfTxRequestMessage struct {
	baseMessage
	TxID string
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlockHashOfTxRequestMessage) Command() MessageCommand {
	return CmdGetIncludingBlockHashOfTxRequestMessage
}

// NewGetIncludingBlockHashOfTxRequest returns a instance of the message
func NewGetIncludingBlockHashOfTxRequest(txID string) *GetIncludingBlockHashOfTxRequestMessage {
	return &GetIncludingBlockHashOfTxRequestMessage{
		TxID: txID,
	}
}

// GetIncludingBlockHashOfTxResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetIncludingBlockHashOfTxResponseMessage struct {
	baseMessage
	Hash string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetIncludingBlockHashOfTxResponseMessage) Command() MessageCommand {
	return CmdGetIncludingBlockHashOfTxResponseMessage
}

// NewGetIncludingBlockHashOfTxResponse returns an instance of the message
func NewGetIncludingBlockHashOfTxResponse(hash string) *GetIncludingBlockHashOfTxResponseMessage {
	return &GetIncludingBlockHashOfTxResponseMessage{
		Hash: hash,
	}
}
