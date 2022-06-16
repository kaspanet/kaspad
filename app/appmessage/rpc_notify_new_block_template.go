package appmessage

// NotifyNewBlockTemplateRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyNewBlockTemplateRequestMessage struct {
	baseMessage
	Id string
}

// Command returns the protocol command string for the message
func (msg *NotifyNewBlockTemplateRequestMessage) Command() MessageCommand {
	return CmdNotifyNewBlockTemplateRequestMessage
}

// NewNotifyNewBlockTemplateRequestMessage returns an instance of the message
func NewNotifyNewBlockTemplateRequestMessage(id string) *NotifyNewBlockTemplateRequestMessage {
	return &NotifyNewBlockTemplateRequestMessage{Id: id}
}

// NotifyNewBlockTemplateResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyNewBlockTemplateResponseMessage struct {
	baseMessage
	Id    string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyNewBlockTemplateResponseMessage) Command() MessageCommand {
	return CmdNotifyNewBlockTemplateResponseMessage
}

// NewNotifyNewBlockTemplateResponseMessage returns an instance of the message
func NewNotifyNewBlockTemplateResponseMessage(id string) *NotifyNewBlockTemplateResponseMessage {
	return &NotifyNewBlockTemplateResponseMessage{Id: id}
}

// NewBlockTemplateNotificationMessage is an appmessage corresponding to
// its respective RPC message
type NewBlockTemplateNotificationMessage struct {
	baseMessage
	Id string
}

// Command returns the protocol command string for the message
func (msg *NewBlockTemplateNotificationMessage) Command() MessageCommand {
	return CmdNewBlockTemplateNotificationMessage
}

// NewNewBlockTemplateNotificationMessage returns an instance of the message
func NewNewBlockTemplateNotificationMessage(id string) *NewBlockTemplateNotificationMessage {
	return &NewBlockTemplateNotificationMessage{Id: id}
}
