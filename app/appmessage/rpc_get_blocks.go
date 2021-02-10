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
func NewGetBlocksRequestMessage(lowHash string, includeBlockHexes bool, includeBlockVerboseData bool) *GetBlocksRequestMessage {
	return &GetBlocksRequestMessage{
		LowHash:                 lowHash,
		IncludeBlockVerboseData: includeBlockVerboseData,
	}
}

// GetBlocksResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlocksResponseMessage struct {
	baseMessage
	BlockHashes      []string
	BlockVerboseData []*BlockVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlocksResponseMessage) Command() MessageCommand {
	return CmdGetBlocksResponseMessage
}

// NewGetBlocksResponseMessage returns a instance of the message
func NewGetBlocksResponseMessage(blockHashes []string, blockHexes []string,
	blockVerboseData []*BlockVerboseData) *GetBlocksResponseMessage {

	return &GetBlocksResponseMessage{
		BlockHashes:      blockHashes,
		BlockVerboseData: blockVerboseData,
	}
}
