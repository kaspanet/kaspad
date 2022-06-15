package appmessage

// NotifyPruningPointUTXOSetOverrideRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyPruningPointUTXOSetOverrideRequestMessage struct {
	baseMessage
	Id string
}

// Command returns the protocol command string for the message
func (msg *NotifyPruningPointUTXOSetOverrideRequestMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideRequestMessage
}

// NewNotifyPruningPointUTXOSetOverrideRequestMessage returns a instance of the message
func NewNotifyPruningPointUTXOSetOverrideRequestMessage(id string) *NotifyPruningPointUTXOSetOverrideRequestMessage {
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
func NewNotifyPruningPointUTXOSetOverrideResponseMessage(id string) *NotifyPruningPointUTXOSetOverrideResponseMessage {
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
func NewPruningPointUTXOSetOverrideNotificationMessage(id string) *PruningPointUTXOSetOverrideNotificationMessage {
	return &PruningPointUTXOSetOverrideNotificationMessage{}
}

// StopNotifyingPruningPointUTXOSetOverrideRequestMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyingPruningPointUTXOSetOverrideRequestMessage struct {
	baseMessage
	Id string
}

// Command returns the protocol command string for the message
func (msg *StopNotifyingPruningPointUTXOSetOverrideRequestMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideRequestMessage
}

// NewStopNotifyingPruningPointUTXOSetOverrideRequestMessage returns a instance of the message
func NewStopNotifyingPruningPointUTXOSetOverrideRequestMessage(id string) *StopNotifyingPruningPointUTXOSetOverrideRequestMessage {
	return &StopNotifyingPruningPointUTXOSetOverrideRequestMessage{Id: id}
}

// StopNotifyingPruningPointUTXOSetOverrideResponseMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyingPruningPointUTXOSetOverrideResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *StopNotifyingPruningPointUTXOSetOverrideResponseMessage) Command() MessageCommand {
	return CmdNotifyPruningPointUTXOSetOverrideResponseMessage
}

// NewStopNotifyingPruningPointUTXOSetOverrideResponseMessage returns a instance of the message
func NewStopNotifyingPruningPointUTXOSetOverrideResponseMessage(id string) *StopNotifyingPruningPointUTXOSetOverrideResponseMessage {
	return &StopNotifyingPruningPointUTXOSetOverrideResponseMessage{}
}
