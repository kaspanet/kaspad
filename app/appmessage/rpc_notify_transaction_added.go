package appmessage

// NotifyTransactionAddedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyTransactionAddedRequestMessage struct {
	baseMessage
	Transaction *MsgTx
}

// Command returns the protocol command string for the message
func (msg *NotifyTransactionAddedRequestMessage) Command() MessageCommand {
	return CmdNotifyTransactionAddedRequestMessage
}

// NewNotifyTransactionAddedRequestMessage returns a instance of the message
func NewNotifyTransactionAddedRequestMessage() *NotifyTransactionAddedRequestMessage {
	return &NotifyTransactionAddedRequestMessage{}
}

// NotifyTransactionAddedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyTransactionAddedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyTransactionAddedResponseMessage) Command() MessageCommand {
	return CmdNotifyTransactionAddedResponseMessage
}

// NewNotifyTransactionAddedResponseMessage returns a instance of the message
func NewNotifyTransactionAddedResponseMessage() *NotifyTransactionAddedResponseMessage {
	return &NotifyTransactionAddedResponseMessage{}
}

// TransactionAddedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type TransactionAddedNotificationMessage struct {
	baseMessage
	Transaction *MsgTx
}

// Command returns the protocol command string for the message
func (msg *TransactionAddedNotificationMessage) Command() MessageCommand {
	return CmdTransactionAddedNotificationMessage
}

// NewTransactionAddedNotificationMessage returns a instance of the message
func NewTransactionAddedNotificationMessage(transaction *MsgTx) *TransactionAddedNotificationMessage {
	return &TransactionAddedNotificationMessage{
		Transaction: transaction,
	}
}
