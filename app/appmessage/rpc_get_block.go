package appmessage

// GetBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockRequestMessage struct {
	baseMessage
	Hash                          string
	SubnetworkID                  string
	IncludeBlockHex               bool
	IncludeBlockVerboseData       bool
	IncludeTransactionVerboseData bool
}

// Command returns the protocol command string for the message
func (msg *GetBlockRequestMessage) Command() MessageCommand {
	return CmdGetBlockRequestMessage
}

// GetBlockRequestMessage returns a instance of the message
func NewGetBlockRequestMessage(hash string, subnetworkID string, includeBlockHex bool,
	includeBlockVerboseData bool, includeTransactionVerboseData bool) *GetBlockRequestMessage {
	return &GetBlockRequestMessage{
		Hash:                          hash,
		SubnetworkID:                  subnetworkID,
		IncludeBlockHex:               includeBlockHex,
		IncludeBlockVerboseData:       includeBlockVerboseData,
		IncludeTransactionVerboseData: includeTransactionVerboseData,
	}
}

// GetBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockResponseMessage struct {
	baseMessage
	BlockHex         string
	BlockVerboseData *BlockVerboseData

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockResponseMessage) Command() MessageCommand {
	return CmdGetBlockResponseMessage
}

// GetBlockResponseMessage returns a instance of the message
func NewGetBlockResponseMessage() *GetBlockResponseMessage {
	return &GetBlockResponseMessage{}
}

type BlockVerboseData struct {
	Hash                   string
	Confirmations          uint64
	Size                   int32
	BlueScore              uint64
	IsChainBlock           bool
	Version                int32
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
	ChildHashes            []string
	AcceptedBlockHashes    []string
}

type TransactionVerboseData struct {
	Hex          string
	TxID         string
	Hash         string
	Size         int32
	Version      int32
	LockTime     uint64
	SubnetworkID string
	Gas          uint64
	PayloadHash  string
	Payload      string
	Vin          []*Vin
	Vout         []*Vout
	BlockHash    string
	AcceptedBy   string
	IsInMempool  bool
	Time         uint64
	BlockTime    uint64
}

type Vin struct {
	TxID      string
	Vout      uint32
	ScriptSig *ScriptSig
	Sequence  uint64
}

type ScriptSig struct {
	Asm string
	Hex string
}

type Vout struct {
	Value        uint64
	N            uint32
	ScriptPubKey *ScriptPubKeyResult
}

type ScriptPubKeyResult struct {
	Asm     string
	Hex     string
	Type    string
	Address string
}
