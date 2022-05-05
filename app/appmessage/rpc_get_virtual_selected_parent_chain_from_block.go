package appmessage

// GetVirtualSelectedParentChainFromBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetVirtualSelectedParentChainFromBlockRequestMessage struct {
	baseMessage
	StartHash                     string
	IncludeAcceptedTransactionIDs bool
}

// Command returns the protocol command string for the message
func (msg *GetVirtualSelectedParentChainFromBlockRequestMessage) Command() MessageCommand {
	return CmdGetVirtualSelectedParentChainFromBlockRequestMessage
}

// NewGetVirtualSelectedParentChainFromBlockRequestMessage returns a instance of the message
func NewGetVirtualSelectedParentChainFromBlockRequestMessage(
	startHash string, includeAcceptedTransactionIDs bool) *GetVirtualSelectedParentChainFromBlockRequestMessage {

	return &GetVirtualSelectedParentChainFromBlockRequestMessage{
		StartHash:                     startHash,
		IncludeAcceptedTransactionIDs: includeAcceptedTransactionIDs,
	}
}

// AcceptedTransactionIDs is a part of the GetVirtualSelectedParentChainFromBlockResponseMessage and
// VirtualSelectedParentChainChangedNotificationMessage appmessages
type AcceptedTransactionIDs struct {
	AcceptingBlockHash     string
	AcceptedTransactionIDs []string
}

// GetVirtualSelectedParentChainFromBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetVirtualSelectedParentChainFromBlockResponseMessage struct {
	baseMessage
	RemovedChainBlockHashes []string
	AddedChainBlockHashes   []string
	AcceptedTransactionIDs  []*AcceptedTransactionIDs

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetVirtualSelectedParentChainFromBlockResponseMessage) Command() MessageCommand {
	return CmdGetVirtualSelectedParentChainFromBlockResponseMessage
}

// NewGetVirtualSelectedParentChainFromBlockResponseMessage returns a instance of the message
func NewGetVirtualSelectedParentChainFromBlockResponseMessage(removedChainBlockHashes,
	addedChainBlockHashes []string, acceptedTransactionIDs []*AcceptedTransactionIDs) *GetVirtualSelectedParentChainFromBlockResponseMessage {

	return &GetVirtualSelectedParentChainFromBlockResponseMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlockHashes:   addedChainBlockHashes,
		AcceptedTransactionIDs:  acceptedTransactionIDs,
	}
}
