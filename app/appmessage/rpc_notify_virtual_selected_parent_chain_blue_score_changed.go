package appmessage

// NotifyVirtualSelectedParentBlueScoreChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentBlueScoreChangedRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentBlueScoreChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentBlueScoreChangedRequestMessage
}

// NewNotifyVirtualSelectedParentBlueScoreChangedRequestMessage returns a instance of the message
func NewNotifyVirtualSelectedParentBlueScoreChangedRequestMessage() *NotifyVirtualSelectedParentBlueScoreChangedRequestMessage {
	return &NotifyVirtualSelectedParentBlueScoreChangedRequestMessage{}
}

// NotifyVirtualSelectedParentBlueScoreChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentBlueScoreChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentBlueScoreChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentBlueScoreChangedResponseMessage
}

// NewNotifyVirtualSelectedParentBlueScoreChangedResponseMessage returns a instance of the message
func NewNotifyVirtualSelectedParentBlueScoreChangedResponseMessage() *NotifyVirtualSelectedParentBlueScoreChangedResponseMessage {
	return &NotifyVirtualSelectedParentBlueScoreChangedResponseMessage{}
}

// VirtualSelectedParentBlueScoreChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type VirtualSelectedParentBlueScoreChangedNotificationMessage struct {
	baseMessage
	VirtualSelectedParentBlueScore uint64
}

// Command returns the protocol command string for the message
func (msg *VirtualSelectedParentBlueScoreChangedNotificationMessage) Command() MessageCommand {
	return CmdVirtualSelectedParentBlueScoreChangedNotificationMessage
}

// NewVirtualSelectedParentBlueScoreChangedNotificationMessage returns a instance of the message
func NewVirtualSelectedParentBlueScoreChangedNotificationMessage(
	virtualSelectedParentBlueScore uint64) *VirtualSelectedParentBlueScoreChangedNotificationMessage {

	return &VirtualSelectedParentBlueScoreChangedNotificationMessage{
		VirtualSelectedParentBlueScore: virtualSelectedParentBlueScore,
	}
}
