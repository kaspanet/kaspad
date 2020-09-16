package appmessage

// GetMempoolEntryRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntryRequestMessage struct {
	baseMessage
	TxID string
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntryRequestMessage) Command() MessageCommand {
	return CmdGetMempoolEntryRequestMessage
}

// NewGetMempoolEntryRequestMessage returns a instance of the message
func NewGetMempoolEntryRequestMessage(txID string) *GetMempoolEntryRequestMessage {
	return &GetMempoolEntryRequestMessage{TxID: txID}
}

// GetMempoolEntryResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntryResponseMessage struct {
	baseMessage
	Fee                    uint64
	TransactionVerboseData *TransactionVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntryResponseMessage) Command() MessageCommand {
	return CmdGetMempoolEntryResponseMessage
}

// NewGetMempoolEntryResponseMessage returns a instance of the message
func NewGetMempoolEntryResponseMessage(fee uint64, transactionVerboseData *TransactionVerboseData) *GetMempoolEntryResponseMessage {
	return &GetMempoolEntryResponseMessage{
		Fee:                    fee,
		TransactionVerboseData: transactionVerboseData,
	}
}
