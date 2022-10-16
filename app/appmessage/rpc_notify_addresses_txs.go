
package appmessage

// NotifyAddressesTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyAddressesTxsRequestMessage struct {
	baseMessage
	Addresses []string
	RequiredConfirmations uint32 
	IncludePending bool
	IncludeSending bool
	IncludeReceiving bool
}

// Command returns the protocol command string for the message
func (msg *NotifyAddressesTxsRequestMessage) Command() MessageCommand {
	return CmdNotifyAddressesTxsRequestMessage
}

// NewNotifyAddressesTxsRequestMessage returns a instance of the message
func NewNotifyAddressesTxsRequestMessage(addresses []string, requiredConfirmations uint32, 
	includePending bool, includeSending bool, includeReceiving bool) *NotifyAddressesTxsRequestMessage {
	return &NotifyAddressesTxsRequestMessage{
		Addresses: addresses,
		RequiredConfirmations:  requiredConfirmations,
		IncludePending: includePending,
		IncludeSending: includeSending,
		IncludeReceiving: includeReceiving,
	}
}

// NotifyAddressesTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyAddressesTxsResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyAddressesTxsResponseMessage) Command() MessageCommand {
	return CmdNotifyAddressesTxsResponseMessage
}

// NewNotifyTXChangedResponseMessage returns a instance of the message
func NewNotifyAddressesTxsResponseMessage() *NotifyAddressesTxsResponseMessage {
	return &NotifyAddressesTxsResponseMessage{}
}

// AddressesTxsNotificationMessage is an appmessage corresponding to
// its respective RPC message
type AddressesTxsNotificationMessage struct {
	baseMessage
	RequiredConfirmations uint32
	Pending	[]*TxEntriesByAddresses
	Confirmed []*TxEntriesByAddresses
	Unconfirmed []string
	
}

// Command returns the protocol command string for the message
func (msg *AddressesTxsNotificationMessage) Command() MessageCommand {
	return CmdAddressesTxsNotificationMessage
}

// NewAddressesTxsNotificationMessage returns a instance of the message
func NewAddressesTxsNotificationMessage(requiredConfirmations uint32, pending []*TxEntriesByAddresses, 
	confirmed []*TxEntriesByAddresses, unconfirmed []string) *AddressesTxsNotificationMessage {
	return &AddressesTxsNotificationMessage{
		RequiredConfirmations: requiredConfirmations,
		Pending:	pending,
		Confirmed:	confirmed,
		Unconfirmed: 	unconfirmed,
	}
}

// TxEntriesByAddresses is an appmessage corresponding to
// its respective RPC message
type TxEntriesByAddresses struct {
	Sending []*TxEntryByAddress
	Reciving []*TxEntryByAddress
}

// TxEntryByAddress is an appmessage corresponding to
// its respective RPC message
type TxEntryByAddress struct {
	Address string
	TxID string
	Confirmations uint32
}