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
	Entry *MempoolEntry

	Error *RPCError
}

// MempoolEntry represents a transaction in the mempool.
type MempoolEntry struct {
	Fee                    uint64
	TransactionVerboseData *RPCTransactionVerboseData
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntryResponseMessage) Command() MessageCommand {
	return CmdGetMempoolEntryResponseMessage
}

// NewGetMempoolEntryResponseMessage returns a instance of the message
func NewGetMempoolEntryResponseMessage(fee uint64, transactionVerboseData *RPCTransactionVerboseData) *GetMempoolEntryResponseMessage {
	return &GetMempoolEntryResponseMessage{
		Entry: &MempoolEntry{
			Fee:                    fee,
			TransactionVerboseData: transactionVerboseData,
		},
	}
}
