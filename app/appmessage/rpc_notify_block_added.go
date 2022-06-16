package appmessage

// NotifyBlockAddedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedRequestMessage struct {
	Id string
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedRequestMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedRequestMessage
}

// NewNotifyBlockAddedRequestMessage returns a instance of the message
func NewNotifyBlockAddedRequestMessage(id string) *NotifyBlockAddedRequestMessage {
	return &NotifyBlockAddedRequestMessage{Id: id}
}

// NotifyBlockAddedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedResponseMessage struct {
	baseMessage
	Id    string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedResponseMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedResponseMessage
}

// NewNotifyBlockAddedResponseMessage returns a instance of the message
func NewNotifyBlockAddedResponseMessage(id string) *NotifyBlockAddedResponseMessage {
	return &NotifyBlockAddedResponseMessage{Id: id}
}

// BlockAddedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type BlockAddedNotificationMessage struct {
	baseMessage
	Id    string
	Block *RPCBlock
}

// Command returns the protocol command string for the message
func (msg *BlockAddedNotificationMessage) Command() MessageCommand {
	return CmdBlockAddedNotificationMessage
}

// NewBlockAddedNotificationMessage returns a instance of the message
func NewBlockAddedNotificationMessage(block *RPCBlock, id string) *BlockAddedNotificationMessage {
	return &BlockAddedNotificationMessage{
		Id:    id,
		Block: block,
	}
}
