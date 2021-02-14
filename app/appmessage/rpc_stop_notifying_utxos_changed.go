package appmessage

// StopNotifyingUTXOsChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyingUTXOsChangedRequestMessage struct {
	baseMessage
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *StopNotifyingUTXOsChangedRequestMessage) Command() MessageCommand {
	return CmdStopNotifyingUTXOsChangedRequestMessage
}

// NewStopNotifyingUTXOsChangedRequestMessage returns a instance of the message
func NewStopNotifyingUTXOsChangedRequestMessage(addresses []string) *StopNotifyingUTXOsChangedRequestMessage {
	return &StopNotifyingUTXOsChangedRequestMessage{
		Addresses: addresses,
	}
}

// StopNotifyingUTXOsChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type StopNotifyingUTXOsChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *StopNotifyingUTXOsChangedResponseMessage) Command() MessageCommand {
	return CmdStopNotifyingUTXOsChangedResponseMessage
}

// NewStopNotifyingUTXOsChangedResponseMessage returns a instance of the message
func NewStopNotifyingUTXOsChangedResponseMessage() *StopNotifyingUTXOsChangedResponseMessage {
	return &StopNotifyingUTXOsChangedResponseMessage{}
}
