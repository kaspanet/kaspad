package appmessage

// NotifyUTXOsChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyUTXOsChangedRequestMessage struct {
	baseMessage
	Addresses []string
}

// Command returns the protocol command string for the message
func (msg *NotifyUTXOsChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyUTXOsChangedRequestMessage
}

// NewNotifyUTXOsChangedRequestMessage returns a instance of the message
func NewNotifyUTXOsChangedRequestMessage(addresses []string) *NotifyUTXOsChangedRequestMessage {
	return &NotifyUTXOsChangedRequestMessage{
		Addresses: addresses,
	}
}

// NotifyUTXOsChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyUTXOsChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyUTXOsChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyUTXOsChangedResponseMessage
}

// NewNotifyUTXOsChangedResponseMessage returns a instance of the message
func NewNotifyUTXOsChangedResponseMessage() *NotifyUTXOsChangedResponseMessage {
	return &NotifyUTXOsChangedResponseMessage{}
}

// UTXOsChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type UTXOsChangedNotificationMessage struct {
	baseMessage
	Added   []*UTXOsByAddressesEntry
	Removed []*UTXOsByAddressesEntry
}

type UTXOsByAddressesEntry struct {
	Address   string
	Outpoint  *RPCOutpoint
	UTXOEntry *RPCUTXOEntry
}

type RPCOutpoint struct {
	TransactionID string
	Index         uint32
}

type RPCUTXOEntry struct {
	Amount         uint64
	ScriptPubKey   string
	BlockBlueScore uint64
	IsCoinbase     bool
}

// Command returns the protocol command string for the message
func (msg *UTXOsChangedNotificationMessage) Command() MessageCommand {
	return CmdUTXOsChangedNotificationMessage
}

// NewUTXOsChangedNotificationMessage returns a instance of the message
func NewUTXOsChangedNotificationMessage() *UTXOsChangedNotificationMessage {
	return &UTXOsChangedNotificationMessage{}
}
