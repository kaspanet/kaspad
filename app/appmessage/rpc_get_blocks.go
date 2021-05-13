package appmessage

// GetBlocksRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlocksRequestMessage struct {
	baseMessage
	LowHash             string
	IncludeBlocks       bool
	IncludeTransactions bool
}

// Command returns the protocol command string for the message
func (msg *GetBlocksRequestMessage) Command() MessageCommand {
	return CmdGetBlocksRequestMessage
}

// NewGetBlocksRequestMessage returns a instance of the message
func NewGetBlocksRequestMessage(lowHash string, includeBlocks bool,
	includeTransactions bool) *GetBlocksRequestMessage {
	return &GetBlocksRequestMessage{
		LowHash:             lowHash,
		IncludeBlocks:       includeBlocks,
		IncludeTransactions: includeTransactions,
	}
}

// GetBlocksResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlocksResponseMessage struct {
	baseMessage
	BlockHashes []string
	Blocks      []*RPCBlock

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlocksResponseMessage) Command() MessageCommand {
	return CmdGetBlocksResponseMessage
}

// NewGetBlocksResponseMessage returns a instance of the message
func NewGetBlocksResponseMessage() *GetBlocksResponseMessage {
	return &GetBlocksResponseMessage{}
}
