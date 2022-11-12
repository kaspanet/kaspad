package appmessage

// ModifyNotifyingAddressesTxsRequestMessage is an appmessage corresponding to
// its respective RPC message
type ModifyNotifyingAddressesTxsRequestMessage struct {
	baseMessage
	AddAddresses          []string
	RemoveAddresses       []string
	RequiredConfirmations uint32
	IncludePending        bool
	IncludeSending        bool
	IncludeReceiving      bool
}

// Command returns the protocol command string for the message
func (msg *ModifyNotifyingAddressesTxsRequestMessage) Command() MessageCommand {
	return CmdModifyNotifyingAddressesTxsRequestMessage
}

// NewModifyNotifyingAddressesTxsRequestMessage returns a instance of the message
func NewModifyNotifyingAddressesTxsRequestMessage(addAddresses []string, removeAddresses []string,
	requiredConfirmations uint32, includePending bool, includeSending bool,
	includeReceiving bool) *ModifyNotifyingAddressesTxsRequestMessage {
	return &ModifyNotifyingAddressesTxsRequestMessage{
		AddAddresses:          addAddresses,
		RemoveAddresses:       removeAddresses,
		RequiredConfirmations: requiredConfirmations,
		IncludePending:        includePending,
		IncludeSending:        includeSending,
		IncludeReceiving:      includeReceiving,
	}
}

// ModifyNotifyingAddressesTxsResponseMessage is an appmessage corresponding to
// its respective RPC message
type ModifyNotifyingAddressesTxsResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *ModifyNotifyingAddressesTxsResponseMessage) Command() MessageCommand {
	return CmdModifyNotifyingAddressesTxsResponseMessage
}

// NewModifyNotifyingAddressesTxsResponseMessage returns a instance of the message
func NewModifyNotifyingAddressesTxsResponseMessage() *NotifyAddressesTxsResponseMessage {
	return &NotifyAddressesTxsResponseMessage{}
}
