package appmessage

// GetAcceptingBlockHashOfTxRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockHashOfTxRequestMessage struct {
	baseMessage
	TxID string
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockHashOfTxRequestMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockHashOfTxRequestMessage
}

// NewGetAcceptingBlockHashOfTxRequest returns a instance of the message
func NewGetAcceptingBlockHashOfTxRequest(txID string) *GetAcceptingBlockHashOfTxRequestMessage {
	return &GetAcceptingBlockHashOfTxRequestMessage{
		TxID: txID,
	}
}

// GetAcceptingBlockHashOfTxResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetAcceptingBlockHashOfTxResponseMessage struct {
	baseMessage
	Hash string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetAcceptingBlockHashOfTxResponseMessage) Command() MessageCommand {
	return CmdGetAcceptingBlockHashOfTxResponseMessage
}

// NewGetAcceptingBlockHashOfTxResponse returns an instance of the message
func NewGetAcceptingBlockHashOfTxResponse(hash string) *GetAcceptingBlockHashOfTxResponseMessage {
	return &GetAcceptingBlockHashOfTxResponseMessage{
		Hash: hash,
	}
}
