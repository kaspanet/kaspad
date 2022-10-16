package appmessage

// NotifyTxsConfirmationChangedRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyTxsConfirmationChangedRequestMessage struct {
	baseMessage
	TxIDs                 []string
	RequiredConfirmations uint32
	IncludePending        bool
}

// Command returns the protocol command string for the message
func (msg *NotifyTxsConfirmationChangedRequestMessage) Command() MessageCommand {
	return CmdNotifyTxsConfirmationChangedRequestMessage
}

// NewNotifyTxsConfirmationChangedRequestMessage returns a instance of the message
func NewNotifyTxsConfirmationChangedRequestMessage(TxIDs []string, requiredConfirmations uint32,
	includePending bool) *NotifyTxsConfirmationChangedRequestMessage {
	return &NotifyTxsConfirmationChangedRequestMessage{
		TxIDs:                 TxIDs,
		RequiredConfirmations: requiredConfirmations,
		IncludePending:        includePending,
	}
}

// NotifyTxsConfirmationChangedResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyTxsConfirmationChangedResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyTxsConfirmationChangedResponseMessage) Command() MessageCommand {
	return CmdNotifyTxsConfirmationChangedResponseMessage
}

// NewNotifyTxsChangedResponseMessage returns a instance of the message
func NewNotifyTxsChangedResponseMessage() *NotifyTxsConfirmationChangedResponseMessage {
	return &NotifyTxsConfirmationChangedResponseMessage{}
}

// TxsConfirmationChangedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type TxsConfirmationChangedNotificationMessage struct {
	baseMessage
	RequiredConfirmations uint32
	Pending               []*TxIDConfirmationsPair
	Confirmed             []*TxIDConfirmationsPair
	UnconfirmedTxIds      []string
}

// Command returns the protocol command string for the message
func (msg *TxsConfirmationChangedNotificationMessage) Command() MessageCommand {
	return CmdTxsConfirmationChangedNotificationMessage
}

// NewTxsChangedNotificationMessage returns a instance of the message
func NewTxsChangedNotificationMessage(requiredConfirmations uint32, pending []*TxIDConfirmationsPair,
	confirmed []*TxIDConfirmationsPair, unconfirmedTxIds []string) *TxsConfirmationChangedNotificationMessage {
	return &TxsConfirmationChangedNotificationMessage{
		RequiredConfirmations: requiredConfirmations,
		Pending:               pending,
		Confirmed:             confirmed,
		UnconfirmedTxIds:      unconfirmedTxIds,
	}
}
