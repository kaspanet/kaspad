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

// RPCTransaction is a kaspad transaction representation meant to be
// used over RPC
type RPCTransaction struct {
	Version      uint16
	Inputs       []*RPCTransactionInput
	Outputs      []*RPCTransactionOutput
	LockTime     uint64
	SubnetworkID string
	Gas          uint64
	PayloadHash  string
	Payload      string
}

// RPCTransactionInput is a kaspad transaction input representation
// meant to be used over RPC
type RPCTransactionInput struct {
	PreviousOutpoint *RPCOutpoint
	SignatureScript  string
	Sequence         uint64
}

// RPCScriptPublicKey is a kaspad ScriptPublicKey representation
type RPCScriptPublicKey struct {
	Version uint16
	Script  string
}

// RPCTransactionOutput is a kaspad transaction output representation
// meant to be used over RPC
type RPCTransactionOutput struct {
	Amount          uint64
	ScriptPublicKey *RPCScriptPublicKey
}

// RPCOutpoint is a kaspad outpoint representation meant to be used
// over RPC
type RPCOutpoint struct {
	TransactionID string
	Index         uint32
}

// RPCUTXOEntry is a kaspad utxo entry representation meant to be used
// over RPC
type RPCUTXOEntry struct {
	Amount          uint64
	ScriptPublicKey *RPCScriptPublicKey
	BlockBlueScore  uint64
	IsCoinbase      bool
}
