package appmessage

type ModifyNotifyingTxsConfirmationChangedRequestMessage struct {
	baseMessage
	AddTxIDs []string
	RemoveTxIDs []string
	RequiredConfirmations uint32 
	IncludePending bool
}

// Command returns the protocol command string for the message
func (msg *ModifyNotifyingTxsConfirmationChangedRequestMessage) Command() MessageCommand {
	return CmdModifyNotifyingTxsConfirmationChangedRequestMessage
}

// NewModifyNotifyingTxsConfirmationChangedRequestMessage returns a instance of the message
func NewModifyNotifyingTxsConfirmationChangedRequestMessage(addTxIDs []string, removeTxIDs []string, 
	requiredConfirmations uint32, includePending bool) *ModifyNotifyingTxsConfirmationChangedRequestMessage {
	return &ModifyNotifyingTxsConfirmationChangedRequestMessage{
		AddTxIDs: addTxIDs,
		RemoveTxIDs: removeTxIDs,
		RequiredConfirmations:  requiredConfirmations,
		IncludePending: includePending,
	}
}

// ModifyNotifyingTxsConfirmationChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type ModifyNotifyingTxsConfirmationChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *ModifyNotifyingTxsConfirmationChangedResponseMessage) Command() MessageCommand {
	return CmdModifyNotifyingTxsConfirmationChangedResponseMessage
}

// NewModifyNotifyingTXChangedResponseMessage returns a instance of the message
func NewModifyNotifyingTxsChangedResponseMessage() *NotifyTxsConfirmationChangedResponseMessage {
	return &NotifyTxsConfirmationChangedResponseMessage{}
}
