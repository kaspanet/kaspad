package appmessage

// NotifyUTXOOfAddressChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyUTXOOfAddressChangedRequestMessage struct {
	baseMessage
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *NotifyUTXOOfAddressChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyUTXOOfAddressChangedRequestMessage
}

// NewNotifyUTXOOfAddressChangedRequestMessage returns a instance of the message
func NewNotifyUTXOOfAddressChangedRequestMessage(addresses []string) *NotifyUTXOOfAddressChangedRequestMessage {
	return &NotifyUTXOOfAddressChangedRequestMessage{
		Addresses: addresses,
	}
}

// NotifyUTXOOfAddressChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyUTXOOfAddressChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyUTXOOfAddressChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyUTXOOfAddressChangedResponseMessage
}

// NewNotifyUTXOOfAddressChangedResponseMessage returns a instance of the message
func NewNotifyUTXOOfAddressChangedResponseMessage() *NotifyUTXOOfAddressChangedResponseMessage {
	return &NotifyUTXOOfAddressChangedResponseMessage{}
}

// UTXOOfAddressChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type UTXOOfAddressChangedNotificationMessage struct {
	baseMessage
	ChangedAddresses []string
}

// Command returns the protocol command string for the message
func (msg *UTXOOfAddressChangedNotificationMessage) Command() MessageCommand {
	return CmdUTXOOfAddressChangedNotificationMessage
}

// NewUTXOOfAddressChangedNotificationMessage returns a instance of the message
func NewUTXOOfAddressChangedNotificationMessage(changedAddresses []string) *UTXOOfAddressChangedNotificationMessage {
	return &UTXOOfAddressChangedNotificationMessage{
		ChangedAddresses: changedAddresses,
	}
}
