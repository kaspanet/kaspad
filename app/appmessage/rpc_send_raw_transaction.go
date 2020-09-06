package appmessage

// SendRawTransactionRequestMessage is an appmessage corresponding to
// its respective RPC message
type SendRawTransactionRequestMessage struct {
	baseMessage
	TransactionHex string
}

// Command returns the protocol command string for the message
func (msg *SendRawTransactionRequestMessage) Command() MessageCommand {
	return CmdSendRawTransactionRequestMessage
}

// NewSendRawTransactionRequestMessage returns a instance of the message
func NewSendRawTransactionRequestMessage(transactionHex string) *SendRawTransactionRequestMessage {
	return &SendRawTransactionRequestMessage{
		TransactionHex: transactionHex,
	}
}

// SendRawTransactionResponseMessage is an appmessage corresponding to
// its respective RPC message
type SendRawTransactionResponseMessage struct {
	baseMessage
	TxID string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *SendRawTransactionResponseMessage) Command() MessageCommand {
	return CmdSendRawTransactionResponseMessage
}

// NewSendRawTransactionResponseMessage returns a instance of the message
func NewSendRawTransactionResponseMessage(txID string) *SendRawTransactionResponseMessage {
	return &SendRawTransactionResponseMessage{
		TxID: txID,
	}
}
