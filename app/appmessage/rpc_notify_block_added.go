package appmessage

// NotifyBlockAddedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedRequestMessage struct {
	baseMessage
	BlockHex string
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedRequestMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedRequestMessage
}

// NotifyBlockAddedRequestMessage returns a instance of the message
func NewNotifyBlockAddedRequestMessage() *NotifyBlockAddedRequestMessage {
	return &NotifyBlockAddedRequestMessage{}
}

// NotifyBlockAddedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyBlockAddedResponseMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyBlockAddedResponseMessage) Command() MessageCommand {
	return CmdNotifyBlockAddedResponseMessage
}

// NotifyBlockAddedResponseMessage returns a instance of the message
func NewNotifyBlockAddedResponseMessage() *NotifyBlockAddedResponseMessage {
	return &NotifyBlockAddedResponseMessage{}
}
