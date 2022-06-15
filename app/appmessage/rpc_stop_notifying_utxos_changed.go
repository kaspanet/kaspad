package appmessage

// StopNotifyingUTXOsChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyingUTXOsChangedRequestMessage struct {
	baseMessage
	Id string
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *StopNotifyingUTXOsChangedRequestMessage) Command() MessageCommand {
	return CmdStopNotifyingUTXOsChangedRequestMessage
}

// NewStopNotifyingUTXOsChangedRequestMessage returns a instance of the message
func NewStopNotifyingUTXOsChangedRequestMessage(addresses []string, id string) *StopNotifyingUTXOsChangedRequestMessage {
	return &StopNotifyingUTXOsChangedRequestMessage{
		Id: id,
		Addresses: addresses,
	}
}

// StopNotifyingUTXOsChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyingUTXOsChangedResponseMessage struct {
	baseMessage
	Id string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *StopNotifyingUTXOsChangedResponseMessage) Command() MessageCommand {
	return CmdStopNotifyingUTXOsChangedResponseMessage
}

// NewStopNotifyingUTXOsChangedResponseMessage returns a instance of the message
func NewStopNotifyingUTXOsChangedResponseMessage(id string) *StopNotifyingUTXOsChangedResponseMessage {
	return &StopNotifyingUTXOsChangedResponseMessage{Id: id}
}
