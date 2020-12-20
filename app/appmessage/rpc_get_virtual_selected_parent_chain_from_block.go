package appmessage

// GetVirtualSelectedParentChainFromBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetVirtualSelectedParentChainFromBlockRequestMessage struct {
	baseMessage
	StartHash string
}

// Command returns the protocol command string for the message
func (msg *GetVirtualSelectedParentChainFromBlockRequestMessage) Command() MessageCommand {
	return CmdGetVirtualSelectedParentChainFromBlockRequestMessage
}

// NewGetVirtualSelectedParentChainFromBlockRequestMessage returns a instance of the message
func NewGetVirtualSelectedParentChainFromBlockRequestMessage(startHash string) *GetVirtualSelectedParentChainFromBlockRequestMessage {
	return &GetVirtualSelectedParentChainFromBlockRequestMessage{
		StartHash: startHash,
	}
}

// GetVirtualSelectedParentChainFromBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetVirtualSelectedParentChainFromBlockResponseMessage struct {
	baseMessage
	RemovedChainBlockHashes []string
	AddedChainBlocks        []*ChainBlock

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetVirtualSelectedParentChainFromBlockResponseMessage) Command() MessageCommand {
	return CmdGetVirtualSelectedParentChainFromBlockResponseMessage
}

// NewGetVirtualSelectedParentChainFromBlockResponseMessage returns a instance of the message
func NewGetVirtualSelectedParentChainFromBlockResponseMessage(removedChainBlockHashes []string,
	addedChainBlocks []*ChainBlock) *GetVirtualSelectedParentChainFromBlockResponseMessage {

	return &GetVirtualSelectedParentChainFromBlockResponseMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}
}
