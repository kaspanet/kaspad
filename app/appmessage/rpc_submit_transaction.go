package appmessage

// SubmitTransactionRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitTransactionRequestMessage struct {
	baseMessage
	Transaction *RPCTransaction
	AllowOrphan bool
}

// Command returns the protocol command string for the message
func (msg *SubmitTransactionRequestMessage) Command() MessageCommand {
	return CmdSubmitTransactionRequestMessage
}

// NewSubmitTransactionRequestMessage returns a instance of the message
func NewSubmitTransactionRequestMessage(transaction *RPCTransaction, allowOrphan bool) *SubmitTransactionRequestMessage {
	return &SubmitTransactionRequestMessage{
		Transaction: transaction,
		AllowOrphan: allowOrphan,
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
	Payload      string
	VerboseData  *RPCTransactionVerboseData
}

// RPCTransactionInput is a kaspad transaction input representation
// meant to be used over RPC
type RPCTransactionInput struct {
	PreviousOutpoint *RPCOutpoint
	SignatureScript  string
	Sequence         uint64
	SigOpCount       byte
	VerboseData      *RPCTransactionInputVerboseData
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
	VerboseData     *RPCTransactionOutputVerboseData
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
	BlockDAAScore   uint64
	IsCoinbase      bool
}

// RPCTransactionVerboseData holds verbose data about a transaction
type RPCTransactionVerboseData struct {
	TransactionID string
	Hash          string
	Mass          uint64
	BlockHash     string
	BlockTime     uint64
}

// RPCTransactionInputVerboseData holds data about a transaction input
type RPCTransactionInputVerboseData struct {
}

// RPCTransactionOutputVerboseData holds data about a transaction output
type RPCTransactionOutputVerboseData struct {
	ScriptPublicKeyType    string
	ScriptPublicKeyAddress string
}
