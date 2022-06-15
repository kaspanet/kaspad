package appmessage

// NotifyVirtualDaaScoreChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualDaaScoreChangedRequestMessage struct {
	baseMessage
	Id string
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualDaaScoreChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyVirtualDaaScoreChangedRequestMessage
}

// NewNotifyVirtualDaaScoreChangedRequestMessage returns a instance of the message
func NewNotifyVirtualDaaScoreChangedRequestMessage(id string) *NotifyVirtualDaaScoreChangedRequestMessage {
	return &NotifyVirtualDaaScoreChangedRequestMessage{Id: id}
}

// NotifyVirtualDaaScoreChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyVirtualDaaScoreChangedResponseMessage struct {
	baseMessage
	Id string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyVirtualDaaScoreChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyVirtualDaaScoreChangedResponseMessage
}

// NewNotifyVirtualDaaScoreChangedResponseMessage returns a instance of the message
func NewNotifyVirtualDaaScoreChangedResponseMessage(id string) *NotifyVirtualDaaScoreChangedResponseMessage {
	return &NotifyVirtualDaaScoreChangedResponseMessage{Id: id}
}

// VirtualDaaScoreChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type VirtualDaaScoreChangedNotificationMessage struct {
	baseMessage
	Id string
	VirtualDaaScore uint64
}

// Command returns the protocol command string for the message
func (msg *VirtualDaaScoreChangedNotificationMessage) Command() MessageCommand {
	return CmdVirtualDaaScoreChangedNotificationMessage
}

// NewVirtualDaaScoreChangedNotificationMessage returns a instance of the message
func NewVirtualDaaScoreChangedNotificationMessage(
	virtualDaaScore uint64, id string) *VirtualDaaScoreChangedNotificationMessage {

	return &VirtualDaaScoreChangedNotificationMessage{
		Id: id,
		VirtualDaaScore: virtualDaaScore,
	}
}
