package appmessage

// NotifyVirtualSelectedParentChainChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentChainChangedRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentChainChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentChainChangedRequestMessage
}

// NewNotifyVirtualSelectedParentChainChangedRequestMessage returns a instance of the message
func NewNotifyVirtualSelectedParentChainChangedRequestMessage() *NotifyVirtualSelectedParentChainChangedRequestMessage {
	return &NotifyVirtualSelectedParentChainChangedRequestMessage{}
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
	AddedChainBlocks        []*ChainBlock
}

// ChainBlock represents a DAG chain-block
type ChainBlock struct {
	Hash           string
	AcceptedBlocks []*AcceptedBlock
}

// AcceptedBlock represents a block accepted into the DAG
type AcceptedBlock struct {
	Hash                   string
	AcceptedTransactionIDs []string
}

// Command returns the protocol command string for the message
func (msg *VirtualSelectedParentChainChangedNotificationMessage) Command() MessageCommand {
	return CmdVirtualSelectedParentChainChangedNotificationMessage
}

// NewVirtualSelectedParentChainChangedNotificationMessage returns a instance of the message
func NewVirtualSelectedParentChainChangedNotificationMessage(removedChainBlockHashes []string,
	addedChainBlocks []*ChainBlock) *VirtualSelectedParentChainChangedNotificationMessage {

	return &VirtualSelectedParentChainChangedNotificationMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}
}
