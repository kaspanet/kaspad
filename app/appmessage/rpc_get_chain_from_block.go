package appmessage

// GetChainFromBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetChainFromBlockRequestMessage struct {
	baseMessage
	StartHash               string
	IncludeBlockVerboseData bool
}

// Command returns the protocol command string for the message
func (msg *GetChainFromBlockRequestMessage) Command() MessageCommand {
	return CmdGetChainFromBlockRequestMessage
}

// NewGetChainFromBlockRequestMessage returns a instance of the message
func NewGetChainFromBlockRequestMessage(startHash string, includeBlockVerboseData bool) *GetChainFromBlockRequestMessage {
	return &GetChainFromBlockRequestMessage{
		StartHash:               startHash,
		IncludeBlockVerboseData: includeBlockVerboseData,
	}
}

// GetChainFromBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetChainFromBlockResponseMessage struct {
	baseMessage
	RemovedChainBlockHashes []string
	AddedChainBlocks        []*ChainChangedChainBlock
	BlockVerboseData        []*BlockVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetChainFromBlockResponseMessage) Command() MessageCommand {
	return CmdGetChainFromBlockResponseMessage
}

// NewGetChainFromBlockResponseMessage returns a instance of the message
func NewGetChainFromBlockResponseMessage(removedChainBlockHashes []string,
	addedChainBlocks []*ChainChangedChainBlock, blockVerboseData []*BlockVerboseData) *GetChainFromBlockResponseMessage {

	return &GetChainFromBlockResponseMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
		BlockVerboseData:        blockVerboseData,
	}
}
