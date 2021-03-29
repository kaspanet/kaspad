package appmessage

// GetBlocksRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlocksRequestMessage struct {
	baseMessage
	LowHash                       string
	IncludeBlockVerboseData       bool
	IncludeTransactionVerboseData bool
}

// Command returns the protocol command string for the message
func (msg *GetBlocksRequestMessage) Command() MessageCommand {
	return CmdGetBlocksRequestMessage
}

// NewGetBlocksRequestMessage returns a instance of the message
func NewGetBlocksRequestMessage(lowHash string, includeBlockVerboseData bool,
	includeTransactionVerboseData bool) *GetBlocksRequestMessage {
	return &GetBlocksRequestMessage{
		LowHash:                       lowHash,
		IncludeBlockVerboseData:       includeBlockVerboseData,
		IncludeTransactionVerboseData: includeTransactionVerboseData,
	}
}

// GetBlocksResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlocksResponseMessage struct {
	baseMessage
	Blocks []*RPCBlock

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlocksResponseMessage) Command() MessageCommand {
	return CmdGetBlocksResponseMessage
}

// NewGetBlocksResponseMessage returns a instance of the message
func NewGetBlocksResponseMessage(blocks []*RPCBlock) *GetBlocksResponseMessage {
	return &GetBlocksResponseMessage{
		Blocks: blocks,
	}
}
