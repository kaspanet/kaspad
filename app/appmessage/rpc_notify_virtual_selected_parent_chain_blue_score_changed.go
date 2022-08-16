package appmessage

// NotifyVirtualSelectedParentBlueScoreChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentBlueScoreChangedRequestMessage struct {
	baseMessage
	ID string
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentBlueScoreChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentBlueScoreChangedRequestMessage
}

// NewNotifyVirtualSelectedParentBlueScoreChangedRequestMessage returns a instance of the message
func NewNotifyVirtualSelectedParentBlueScoreChangedRequestMessage(id string) *NotifyVirtualSelectedParentBlueScoreChangedRequestMessage {
	return &NotifyVirtualSelectedParentBlueScoreChangedRequestMessage{ID: id}
}

// NotifyVirtualSelectedParentBlueScoreChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualSelectedParentBlueScoreChangedResponseMessage struct {
	baseMessage
	ID    string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualSelectedParentBlueScoreChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyVirtualSelectedParentBlueScoreChangedResponseMessage
}

// NewNotifyVirtualSelectedParentBlueScoreChangedResponseMessage returns a instance of the message
func NewNotifyVirtualSelectedParentBlueScoreChangedResponseMessage(id string) *NotifyVirtualSelectedParentBlueScoreChangedResponseMessage {
	return &NotifyVirtualSelectedParentBlueScoreChangedResponseMessage{ID: id}
}

// VirtualSelectedParentBlueScoreChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type VirtualSelectedParentBlueScoreChangedNotificationMessage struct {
	baseMessage
	ID                             string
	VirtualSelectedParentBlueScore uint64
}

// Command returns the protocol command string for the message
func (msg *VirtualSelectedParentBlueScoreChangedNotificationMessage) Command() MessageCommand {
	return CmdVirtualSelectedParentBlueScoreChangedNotificationMessage
}

// NewVirtualSelectedParentBlueScoreChangedNotificationMessage returns a instance of the message
func NewVirtualSelectedParentBlueScoreChangedNotificationMessage(
	virtualSelectedParentBlueScore uint64, id string) *VirtualSelectedParentBlueScoreChangedNotificationMessage {

	return &VirtualSelectedParentBlueScoreChangedNotificationMessage{
		ID:                             id,
		VirtualSelectedParentBlueScore: virtualSelectedParentBlueScore,
	}
}
