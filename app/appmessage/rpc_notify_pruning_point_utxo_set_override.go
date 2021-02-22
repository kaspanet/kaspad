package appmessage

// NotifyPruningPointUTXOSetOverrideRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyPruningPointUTXOSetOverrideRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyPruningPointUTXOSetOverrideRequestMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideRequestMessage
}

// NewNotifyPruningPointUTXOSetOverrideRequestMessage returns a instance of the message
func NewNotifyPruningPointUTXOSetOverrideRequestMessage() *NotifyPruningPointUTXOSetOverrideRequestMessage {
	return &NotifyPruningPointUTXOSetOverrideRequestMessage{}
}

// NotifyPruningPointUTXOSetOverrideResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyPruningPointUTXOSetOverrideResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyPruningPointUTXOSetOverrideResponseMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideResponseMessage
}

// NewNotifyPruningPointUTXOSetOverrideResponseMessage returns a instance of the message
func NewNotifyPruningPointUTXOSetOverrideResponseMessage() *NotifyPruningPointUTXOSetOverrideResponseMessage {
	return &NotifyPruningPointUTXOSetOverrideResponseMessage{}
}

// PruningPointUTXOSetOverrideNotificationMessage is an appmessage corresponding to
// its respective RPC message
type PruningPointUTXOSetOverrideNotificationMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *PruningPointUTXOSetOverrideNotificationMessage) Command() MessageCommand {
	return CmdPruningPointUTXOSetOverrideNotificationMessage
}

// NewPruningPointUTXOSetOverrideNotificationMessage returns a instance of the message
func NewPruningPointUTXOSetOverrideNotificationMessage() *PruningPointUTXOSetOverrideNotificationMessage {
	return &PruningPointUTXOSetOverrideNotificationMessage{}
}

// StopNotifyPruningPointUTXOSetOverrideRequestMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyPruningPointUTXOSetOverrideRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *StopNotifyPruningPointUTXOSetOverrideRequestMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideRequestMessage
}

// NewStopNotifyPruningPointUTXOSetOverrideRequestMessage returns a instance of the message
func NewStopNotifyPruningPointUTXOSetOverrideRequestMessage() *StopNotifyPruningPointUTXOSetOverrideRequestMessage {
	return &StopNotifyPruningPointUTXOSetOverrideRequestMessage{}
}

// StopNotifyPruningPointUTXOSetOverrideResponseMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyPruningPointUTXOSetOverrideResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *StopNotifyPruningPointUTXOSetOverrideResponseMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideResponseMessage
}

// NewStopNotifyPruningPointUTXOSetOverrideResponseMessage returns a instance of the message
func NewStopNotifyPruningPointUTXOSetOverrideResponseMessage() *StopNotifyPruningPointUTXOSetOverrideResponseMessage {
	return &StopNotifyPruningPointUTXOSetOverrideResponseMessage{}
}
