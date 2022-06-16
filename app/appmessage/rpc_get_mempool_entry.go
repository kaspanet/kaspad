package appmessage

// GetMempoolEntryRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntryRequestMessage struct {
	baseMessage
	TxID                  string
	IncludeOrphanPool     bool
	FilterTransactionPool bool
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntryRequestMessage) Command() MessageCommand {
	return CmdGetMempoolEntryRequestMessage
}

// NewGetMempoolEntryRequestMessage returns a instance of the message
func NewGetMempoolEntryRequestMessage(txID string, includeOrphanPool bool, filterTransactionPool bool) *GetMempoolEntryRequestMessage {
	return &GetMempoolEntryRequestMessage{
		TxID:                  txID,
		IncludeOrphanPool:     includeOrphanPool,
		FilterTransactionPool: filterTransactionPool,
	}
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
	Fee         uint64
	Transaction *RPCTransaction
	IsOrphan    bool
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntryResponseMessage) Command() MessageCommand {
	return CmdGetMempoolEntryResponseMessage
}

// NewGetMempoolEntryResponseMessage returns a instance of the message
func NewGetMempoolEntryResponseMessage(fee uint64, transaction *RPCTransaction, isOrphan bool) *GetMempoolEntryResponseMessage {
	return &GetMempoolEntryResponseMessage{
		Entry: &MempoolEntry{
			Fee:         fee,
			Transaction: transaction,
			IsOrphan:    isOrphan,
		},
	}
}
