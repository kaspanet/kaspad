package appmessage

// NotifyUTXOsChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyUTXOsChangedRequestMessage struct {
	baseMessage
	ID        string
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *NotifyUTXOsChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyUTXOsChangedRequestMessage
}

// NewNotifyUTXOsChangedRequestMessage returns a instance of the message
func NewNotifyUTXOsChangedRequestMessage(addresses []string, id string) *NotifyUTXOsChangedRequestMessage {
	return &NotifyUTXOsChangedRequestMessage{
		ID:        id,
		Addresses: addresses,
	}
}

// NotifyUTXOsChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyUTXOsChangedResponseMessage struct {
	baseMessage
	ID    string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyUTXOsChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyUTXOsChangedResponseMessage
}

// NewNotifyUTXOsChangedResponseMessage returns a instance of the message
func NewNotifyUTXOsChangedResponseMessage(id string) *NotifyUTXOsChangedResponseMessage {
	return &NotifyUTXOsChangedResponseMessage{ID: id}
}

// UTXOsChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type UTXOsChangedNotificationMessage struct {
	baseMessage
	ID      string
	Added   []*UTXOsByAddressesEntry
	Removed []*UTXOsByAddressesEntry
}

// UTXOsByAddressesEntry represents a UTXO of some address
type UTXOsByAddressesEntry struct {
	Address   string
	Outpoint  *RPCOutpoint
	UTXOEntry *RPCUTXOEntry
}

// Command returns the protocol command string for the message
func (msg *UTXOsChangedNotificationMessage) Command() MessageCommand {
	return CmdUTXOsChangedNotificationMessage
}

// NewUTXOsChangedNotificationMessage returns a instance of the message
func NewUTXOsChangedNotificationMessage(id string) *UTXOsChangedNotificationMessage {
	return &UTXOsChangedNotificationMessage{ID: id}
}
