package appmessage

// GetMempoolEntriesRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntriesRequestMessage struct {
	baseMessage
	IncludeOrphanPool     bool
	FilterTransactionPool bool
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntriesRequestMessage) Command() MessageCommand {
	return CmdGetMempoolEntriesRequestMessage
}

// NewGetMempoolEntriesRequestMessage returns a instance of the message
func NewGetMempoolEntriesRequestMessage(includeOrphanPool bool, filterTransactionPool bool) *GetMempoolEntriesRequestMessage {
	return &GetMempoolEntriesRequestMessage{
		IncludeOrphanPool:     includeOrphanPool,
		FilterTransactionPool: filterTransactionPool,
	}
}

// GetMempoolEntriesResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetMempoolEntriesResponseMessage struct {
	baseMessage
	Entries []*MempoolEntry

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetMempoolEntriesResponseMessage) Command() MessageCommand {
	return CmdGetMempoolEntriesResponseMessage
}

// NewGetMempoolEntriesResponseMessage returns a instance of the message
func NewGetMempoolEntriesResponseMessage(entries []*MempoolEntry) *GetMempoolEntriesResponseMessage {
	return &GetMempoolEntriesResponseMessage{
		Entries: entries,
	}
}
