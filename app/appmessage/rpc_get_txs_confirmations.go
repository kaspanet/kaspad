package appmessage

// TxIDConfirmationsPair is an appmessage corresponding to
// its respective RPC message
type TxIDConfirmationsPair struct {
	TxID          string
	Confirmations int64
}

// GetTxsConfirmationsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetTxsConfirmationsRequestMessage struct {
	baseMessage
	TxIDs []string
}

// Command returns the protocol command string for the message
func (msg *GetTxsConfirmationsRequestMessage) Command() MessageCommand {
	return CmdGetTxsConfirmationsRequestMessage
}

// NewGetTxsConfirmationsRequest returns a instance of the message
func NewGetTxsConfirmationsRequest(txIDs []string) *GetTxsConfirmationsRequestMessage {
	return &GetTxsConfirmationsRequestMessage{
		TxIDs: txIDs,
	}
}

// GetTxsConfirmationsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetTxsConfirmationsResponseMessage struct {
	baseMessage
	TxIDConfirmationsPairs []*TxIDConfirmationsPair

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetTxsConfirmationsResponseMessage) Command() MessageCommand {
	return CmdGetTxsConfirmationsResponseMessage
}

// NewGetTxsConfirmationsResponse returns an instance of the message
func NewGetTxsConfirmationsResponse(txIDConfirmationsPairs []*TxIDConfirmationsPair) *GetTxsConfirmationsResponseMessage {
	return &GetTxsConfirmationsResponseMessage{
		TxIDConfirmationsPairs: txIDConfirmationsPairs,
	}
}
