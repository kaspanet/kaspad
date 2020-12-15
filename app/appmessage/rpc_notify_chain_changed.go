package appmessage

// NotifyChainChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyChainChangedRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyChainChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyChainChangedRequestMessage
}

// NewNotifyChainChangedRequestMessage returns a instance of the message
func NewNotifyChainChangedRequestMessage() *NotifyChainChangedRequestMessage {
	return &NotifyChainChangedRequestMessage{}
}

// NotifyChainChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyChainChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyChainChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyChainChangedResponseMessage
}

// NewNotifyChainChangedResponseMessage returns a instance of the message
func NewNotifyChainChangedResponseMessage() *NotifyChainChangedResponseMessage {
	return &NotifyChainChangedResponseMessage{}
}

// ChainChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type ChainChangedNotificationMessage struct {
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
func (msg *ChainChangedNotificationMessage) Command() MessageCommand {
	return CmdChainChangedNotificationMessage
}

// NewChainChangedNotificationMessage returns a instance of the message
func NewChainChangedNotificationMessage(removedChainBlockHashes []string,
	addedChainBlocks []*ChainBlock) *ChainChangedNotificationMessage {

	return &ChainChangedNotificationMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}
}
