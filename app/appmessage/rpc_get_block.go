package appmessage

// GetBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockRequestMessage struct {
	baseMessage
	Hash                string
	IncludeTransactions bool
}

// Command returns the protocol command string for the message
func (msg *GetBlockRequestMessage) Command() MessageCommand {
	return CmdGetBlockRequestMessage
}

// NewGetBlockRequestMessage returns a instance of the message
func NewGetBlockRequestMessage(hash string, includeTransactions bool) *GetBlockRequestMessage {
	return &GetBlockRequestMessage{
		Hash:                hash,
		IncludeTransactions: includeTransactions,
	}
}

// GetBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockResponseMessage struct {
	baseMessage
	Block *RPCBlock

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockResponseMessage) Command() MessageCommand {
	return CmdGetBlockResponseMessage
}

// NewGetBlockResponseMessage returns a instance of the message
func NewGetBlockResponseMessage() *GetBlockResponseMessage {
	return &GetBlockResponseMessage{}
}
