package appmessage

// GetTxConfirmationsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetTxConfirmationsRequestMessage struct {
	baseMessage
	TxID string
}

// Command returns the protocol command string for the message
func (msg *GetTxConfirmationsRequestMessage) Command() MessageCommand {
	return CmdGetTxConfirmationsRequestMessage
}

// NewGetTxConfirmationsRequest returns a instance of the message
func NewGetTxConfirmationsRequest(txID string) *GetTxConfirmationsRequestMessage {
	return &GetTxConfirmationsRequestMessage{
		TxID: txID,
	}
}

// GetTxConfirmationsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetTxConfirmationsResponseMessage struct {
	baseMessage
	Confirmations int64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetTxConfirmationsResponseMessage) Command() MessageCommand {
	return CmdGetTxConfirmationsResponseMessage
}

// NewGetTxConfirmationsResponse returns an instance of the message
func NewGetTxConfirmationsResponse(confirmations int64) *GetTxConfirmationsResponseMessage {
	return &GetTxConfirmationsResponseMessage{
		Confirmations: confirmations,
	}
}
