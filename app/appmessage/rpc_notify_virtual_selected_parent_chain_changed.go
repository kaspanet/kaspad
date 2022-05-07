package appmessage

// NotifyVirtualSelectedParentChainChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentChainChangedRequestMessage struct {
	baseMessage
	IncludeAcceptedTransactionIDs bool
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentChainChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentChainChangedRequestMessage
}

// NewNotifyVirtualSelectedParentChainChangedRequestMessage returns an instance of the message
func NewNotifyVirtualSelectedParentChainChangedRequestMessage(
	includeAcceptedTransactionIDs bool) *NotifyVirtualSelectedParentChainChangedRequestMessage {

	return &NotifyVirtualSelectedParentChainChangedRequestMessage{
		IncludeAcceptedTransactionIDs: includeAcceptedTransactionIDs,
	}
}

// NotifyVirtualSelectedParentChainChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentChainChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentChainChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentChainChangedResponseMessage
}

// NewNotifyVirtualSelectedParentChainChangedResponseMessage returns a instance of the message
func NewNotifyVirtualSelectedParentChainChangedResponseMessage() *NotifyVirtualSelectedParentChainChangedResponseMessage {
	return &NotifyVirtualSelectedParentChainChangedResponseMessage{}
}

// VirtualSelectedParentChainChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type VirtualSelectedParentChainChangedNotificationMessage struct {
	baseMessage
	RemovedChainBlockHashes []string
	AddedChainBlockHashes   []string
	AcceptedTransactionIDs  []*AcceptedTransactionIDs
}

// Command returns the protocol command string for the message
func (msg *VirtualSelectedParentChainChangedNotificationMessage) Command() MessageCommand {
	return CmdVirtualSelectedParentChainChangedNotificationMessage
}

// NewVirtualSelectedParentChainChangedNotificationMessage returns a instance of the message
func NewVirtualSelectedParentChainChangedNotificationMessage(removedChainBlockHashes,
	addedChainBlocks []string, acceptedTransactionIDs []*AcceptedTransactionIDs) *VirtualSelectedParentChainChangedNotificationMessage {

	return &VirtualSelectedParentChainChangedNotificationMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlockHashes:   addedChainBlocks,
		AcceptedTransactionIDs:  acceptedTransactionIDs,
	}
}
