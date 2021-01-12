package appmessage

// GetBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockRequestMessage struct {
	baseMessage
	Hash                          string
	SubnetworkID                  string
	IncludeTransactionVerboseData bool
}

// Command returns the protocol command string for the message
func (msg *GetBlockRequestMessage) Command() MessageCommand {
	return CmdGetBlockRequestMessage
}

// NewGetBlockRequestMessage returns a instance of the message
func NewGetBlockRequestMessage(hash string, subnetworkID string, includeTransactionVerboseData bool) *GetBlockRequestMessage {
	return &GetBlockRequestMessage{
		Hash:                          hash,
		SubnetworkID:                  subnetworkID,
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
	Version                uint16
	VersionHex             string
	HashMerkleRoot         string
	AcceptedIDMerkleRoot   string
	UTXOCommitment         string
	TxIDs                  []string
	TransactionVerboseData []*TransactionVerboseData
	Time                   int64
	Nonce                  uint64
	Bits                   string
	Difficulty             float64
	ParentHashes           []string
	SelectedParentHash     string
	BlueScore              uint64
	IsHeaderOnly           bool
}

// TransactionVerboseData holds verbose data about a transaction
type TransactionVerboseData struct {
	TxID                      string
	Hash                      string
	Size                      uint64
	Version                   uint16
	LockTime                  uint64
	SubnetworkID              string
	Gas                       uint64
	PayloadHash               string
	Payload                   string
	TransactionVerboseInputs  []*TransactionVerboseInput
	TransactionVerboseOutputs []*TransactionVerboseOutput
	BlockHash                 string
	Time                      uint64
	BlockTime                 uint64
}

// TransactionVerboseInput holds data about a transaction input
type TransactionVerboseInput struct {
	TxID        string
	OutputIndex uint32
	ScriptSig   *ScriptSig
	Sequence    uint64
}

// ScriptSig holds data about a script signature
type ScriptSig struct {
	Asm string
	Hex string
}

// TransactionVerboseOutput holds data about a transaction output
type TransactionVerboseOutput struct {
	Value        uint64
	Index        uint32
	ScriptPubKey *ScriptPubKeyResult
}

// ScriptPubKeyResult holds data about a script public key
type ScriptPubKeyResult struct {
	Hex     string
	Type    string
	Address string
}
