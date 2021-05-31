package appmessage

// NotifyVirtualDaaScoreChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualDaaScoreChangedRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualDaaScoreChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyVirtualDaaScoreChangedRequestMessage
}

// NewNotifyVirtualDaaScoreChangedRequestMessage returns a instance of the message
func NewNotifyVirtualDaaScoreChangedRequestMessage() *NotifyVirtualDaaScoreChangedRequestMessage {
	return &NotifyVirtualDaaScoreChangedRequestMessage{}
}

// NotifyVirtualDaaScoreChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualDaaScoreChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualDaaScoreChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyVirtualDaaScoreChangedResponseMessage
}

// NewNotifyVirtualDaaScoreChangedResponseMessage returns a instance of the message
func NewNotifyVirtualDaaScoreChangedResponseMessage() *NotifyVirtualDaaScoreChangedResponseMessage {
	return &NotifyVirtualDaaScoreChangedResponseMessage{}
}

// VirtualDaaScoreChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type VirtualDaaScoreChangedNotificationMessage struct {
	baseMessage
	VirtualDaaScore uint64
}

// Command returns the protocol command string for the message
func (msg *VirtualDaaScoreChangedNotificationMessage) Command() MessageCommand {
	return CmdVirtualDaaScoreChangedNotificationMessage
}

// NewVirtualDaaScoreChangedNotificationMessage returns a instance of the message
func NewVirtualDaaScoreChangedNotificationMessage(
	virtualDaaScore uint64) *VirtualDaaScoreChangedNotificationMessage {

	return &VirtualDaaScoreChangedNotificationMessage{
		VirtualDaaScore: virtualDaaScore,
	}
}
