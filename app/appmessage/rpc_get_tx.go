package appmessage

// GetTxRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetTxRequestMessage struct {
	baseMessage
	TxID string
}

// Command returns the protocol command string for the message
func (msg *GetTxRequestMessage) Command() MessageCommand {
	return CmdGetTxRequestMessage
}

// NewTxRequest returns a instance of the message
func NewGetTxRequest(txID string) *GetTxRequestMessage {
	return &GetTxRequestMessage{
		TxID: txID,
	}
}

// GetTxResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetTxResponseMessage struct {
	baseMessage
	Transaction *RPCTransaction

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetTxResponseMessage) Command() MessageCommand {
	return CmdGetTxResponseMessage
}

// NewGetTxResponse returns an instance of the message
func NewGetTxResponse(transaction *RPCTransaction) *GetTxResponseMessage {
	return &GetTxResponseMessage{
		Transaction: transaction,
	}
}
