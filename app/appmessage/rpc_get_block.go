package appmessage

// GetBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockRequestMessage struct {
	baseMessage
	Hash                          string
	IncludeTransactionVerboseData bool
}

// Command returns the protocol command string for the message
func (msg *GetBlockRequestMessage) Command() MessageCommand {
	return CmdGetBlockRequestMessage
}

// NewGetBlockRequestMessage returns a instance of the message
func NewGetBlockRequestMessage(hash string, includeTransactionVerboseData bool) *GetBlockRequestMessage {
	return &GetBlockRequestMessage{
		Hash:                          hash,
		IncludeTransactionVerboseData: includeTransactionVerboseData,
	}
}

// GetBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockResponseMessage struct {
	baseMessage
	BlockVerboseData *BlockVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockResponseMessage) Command() MessageCommand {
	return CmdGetBlockResponseMessage
}

// NewGetBlockResponseMessage returns a instance of the message
func NewGetBlockResponseMessage() *GetBlockResponseMessage {
	return &GetBlockResponseMessage{}
}

// BlockVerboseData holds verbose data about a block
type BlockVerboseData struct {
	Hash                   string
	Block                  *RPCBlock
	TxIDs                  []string
	TransactionVerboseData []*TransactionVerboseData
	Difficulty             float64
	ChildrenHashes         []string
	SelectedParentHash     string
	BlueScore              uint64
	IsHeaderOnly           bool
}

// TransactionVerboseData holds verbose data about a transaction
type TransactionVerboseData struct {
	TxID                      string
	Hash                      string
	Size                      uint64
	TransactionVerboseInputs  []*TransactionVerboseInput
	TransactionVerboseOutputs []*TransactionVerboseOutput
	BlockHash                 string
	BlockTime                 uint64
	Transaction               *RPCTransaction
}

// TransactionVerboseInput holds data about a transaction input
type TransactionVerboseInput struct {
}

// TransactionVerboseOutput holds data about a transaction output
type TransactionVerboseOutput struct {
	ScriptPublicKeyType    string
	ScriptPublicKeyAddress string
}
