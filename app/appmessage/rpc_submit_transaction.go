package appmessage

// SubmitTransactionRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitTransactionRequestMessage struct {
	baseMessage
	Transaction *RPCTransaction
}

// Command returns the protocol command string for the message
func (msg *SubmitTransactionRequestMessage) Command() MessageCommand {
	return CmdSubmitTransactionRequestMessage
}

// NewSubmitTransactionRequestMessage returns a instance of the message
func NewSubmitTransactionRequestMessage(transaction *RPCTransaction) *SubmitTransactionRequestMessage {
	return &SubmitTransactionRequestMessage{
		Transaction: transaction,
	}
}

// SubmitTransactionResponseMessage is an appmessage corresponding to
// its respective RPC message
type SubmitTransactionResponseMessage struct {
	baseMessage
	TransactionID string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *SubmitTransactionResponseMessage) Command() MessageCommand {
	return CmdSubmitTransactionResponseMessage
}

// NewSubmitTransactionResponseMessage returns a instance of the message
func NewSubmitTransactionResponseMessage(transactionID string) *SubmitTransactionResponseMessage {
	return &SubmitTransactionResponseMessage{
		TransactionID: transactionID,
	}
}

type RPCTransaction struct {
	Version      int32
	Inputs       []*RPCTransactionInput
	Outputs      []*RPCTransactionOutput
	LockTime     uint64
	SubnetworkID string
	Gas          uint64
	PayloadHash  string
	Payload      string
}

type RPCTransactionInput struct {
	PreviousOutpoint *RPCOutpoint
	SignatureScript  string
	Sequence         uint64
}

type RPCTransactionOutput struct {
	Amount       uint64
	ScriptPubKey string
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
