package appmessage

// NotifyBlockAddedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedRequestMessage struct {
	ID string
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedRequestMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedRequestMessage
}

// NewNotifyBlockAddedRequestMessage returns a instance of the message
func NewNotifyBlockAddedRequestMessage(id string) *NotifyBlockAddedRequestMessage {
	return &NotifyBlockAddedRequestMessage{ID: id}
}

// NotifyBlockAddedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedResponseMessage struct {
	baseMessage
	ID    string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedResponseMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedResponseMessage
}

// NewNotifyBlockAddedResponseMessage returns a instance of the message
func NewNotifyBlockAddedResponseMessage(id string) *NotifyBlockAddedResponseMessage {
	return &NotifyBlockAddedResponseMessage{ID: id}
}

// BlockAddedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type BlockAddedNotificationMessage struct {
	baseMessage
	ID    string
	Block *RPCBlock
}

// Command returns the protocol command string for the message
func (msg *BlockAddedNotificationMessage) Command() MessageCommand {
	return CmdBlockAddedNotificationMessage
}

// NewBlockAddedNotificationMessage returns a instance of the message
func NewBlockAddedNotificationMessage(block *RPCBlock, id string) *BlockAddedNotificationMessage {
	return &BlockAddedNotificationMessage{
		ID:    id,
		Block: block,
	}
}
