package appmessage

// NotifyBlockAddedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedRequestMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedRequestMessage
}

// NewNotifyBlockAddedRequestMessage returns a instance of the message
func NewNotifyBlockAddedRequestMessage() *NotifyBlockAddedRequestMessage {
	return &NotifyBlockAddedRequestMessage{}
}

// NotifyBlockAddedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedResponseMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedResponseMessage
}

// NewNotifyBlockAddedResponseMessage returns a instance of the message
func NewNotifyBlockAddedResponseMessage() *NotifyBlockAddedResponseMessage {
	return &NotifyBlockAddedResponseMessage{}
}

// BlockAddedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type BlockAddedNotificationMessage struct {
	baseMessage
	BlockVerboseData *RPCBlockVerboseData
}

// Command returns the protocol command string for the message
func (msg *BlockAddedNotificationMessage) Command() MessageCommand {
	return CmdBlockAddedNotificationMessage
}

// NewBlockAddedNotificationMessage returns a instance of the message
func NewBlockAddedNotificationMessage(blockVerboseData *RPCBlockVerboseData) *BlockAddedNotificationMessage {
	return &BlockAddedNotificationMessage{
		BlockVerboseData: blockVerboseData,
	}
}
