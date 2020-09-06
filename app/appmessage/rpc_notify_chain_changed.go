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
	AddedChainBlocks        []*ChainChangedChainBlock
}

// ChainChangedChainBlock represents a DAG chain-block
type ChainChangedChainBlock struct {
	Hash           string
	AcceptedBlocks []*ChainChangedAcceptedBlock
}

// ChainChangedAcceptedBlock represents a block accepted into the DAG
type ChainChangedAcceptedBlock struct {
	Hash          string
	AcceptedTxIDs []string
}

// Command returns the protocol command string for the message
func (msg *ChainChangedNotificationMessage) Command() MessageCommand {
	return CmdChainChangedNotificationMessage
}

// NewChainChangedNotificationMessage returns a instance of the message
func NewChainChangedNotificationMessage(removedChainBlockHashes []string,
	addedChainBlocks []*ChainChangedChainBlock) *ChainChangedNotificationMessage {

	return &ChainChangedNotificationMessage{
		RemovedChainBlockHashes: removedChainBlockHashes,
		AddedChainBlocks:        addedChainBlocks,
	}
}
