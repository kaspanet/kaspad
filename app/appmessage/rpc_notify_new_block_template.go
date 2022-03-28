package appmessage

// NotifyNewBlockTemplateRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyNewBlockTemplateRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyNewBlockTemplateRequestMessage) Command() MessageCommand {
	return CmdNotifyNewBlockTemplateRequestMessage
}

// NewNotifyNewBlockTemplateRequestMessage returns an instance of the message
func NewNotifyNewBlockTemplateRequestMessage() *NotifyNewBlockTemplateRequestMessage {
	return &NotifyNewBlockTemplateRequestMessage{}
}

// NotifyNewBlockTemplateResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyNewBlockTemplateResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyNewBlockTemplateResponseMessage) Command() MessageCommand {
	return CmdNotifyNewBlockTemplateResponseMessage
}

// NewNotifyNewBlockTemplateResponseMessage returns an instance of the message
func NewNotifyNewBlockTemplateResponseMessage() *NotifyNewBlockTemplateResponseMessage {
	return &NotifyNewBlockTemplateResponseMessage{}
}

// NewBlockTemplateNotificationMessage is an appmessage corresponding to
// its respective RPC message
type NewBlockTemplateNotificationMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NewBlockTemplateNotificationMessage) Command() MessageCommand {
	return CmdNewBlockTemplateNotificationMessage
}

// NewNewBlockTemplateNotificationMessage returns an instance of the message
func NewNewBlockTemplateNotificationMessage() *NewBlockTemplateNotificationMessage {
	return &NewBlockTemplateNotificationMessage{}
}
