package appmessage

// SubmitBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockRequestMessage struct {
	baseMessage
	Block             *RPCBlock
	AllowNonDAABlocks bool
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockRequestMessage) Command() MessageCommand {
	return CmdSubmitBlockRequestMessage
}

// NewSubmitBlockRequestMessage returns a instance of the message
func NewSubmitBlockRequestMessage(block *RPCBlock, allowNonDAABlocks bool) *SubmitBlockRequestMessage {
	return &SubmitBlockRequestMessage{
		Block:             block,
		AllowNonDAABlocks: allowNonDAABlocks,
	}
}

// RejectReason describes the reason why a block sent by SubmitBlock was rejected
type RejectReason byte

// RejectReason constants
// Not using iota, since in the .proto file those are hardcoded
const (
	RejectReasonNone         RejectReason = 0
	RejectReasonBlockInvalid RejectReason = 1
	RejectReasonIsInIBD      RejectReason = 2
)

var rejectReasonToString = map[RejectReason]string{
	RejectReasonNone:         "None",
	RejectReasonBlockInvalid: "Block is invalid",
	RejectReasonIsInIBD:      "Node is in IBD",
}

func (rr RejectReason) String() string {
	return rejectReasonToString[rr]
}

// SubmitBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockResponseMessage struct {
	baseMessage
	RejectReason RejectReason
	Error        *RPCError
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockResponseMessage) Command() MessageCommand {
	return CmdSubmitBlockResponseMessage
}

// NewSubmitBlockResponseMessage returns an instance of the message
func NewSubmitBlockResponseMessage() *SubmitBlockResponseMessage {
	return &SubmitBlockResponseMessage{}
}

// RPCBlock is a kaspad block representation meant to be
// used over RPC
type RPCBlock struct {
	Header       *RPCBlockHeader
	Transactions []*RPCTransaction
	VerboseData  *RPCBlockVerboseData
}

// RPCBlockHeader is a kaspad block header representation meant to be
// used over RPC
type RPCBlockHeader struct {
	Version              uint32
	Parents              []*RPCBlockLevelParents
	HashMerkleRoot       string
	AcceptedIDMerkleRoot string
	UTXOCommitment       string
	Timestamp            int64
	Bits                 uint32
	Nonce                uint64
	DAAScore             uint64
	BlueScore            uint64
	BlueWork             string
	PruningPoint         string
}

// RPCBlockLevelParents holds parent hashes for one block level
type RPCBlockLevelParents struct {
	ParentHashes []string
}

// RPCBlockVerboseData holds verbose data about a block
type RPCBlockVerboseData struct {
	Hash               string
	Difficulty         float64
	SelectedParentHash string
	TransactionIDs     []string
	IsHeaderOnly       bool
	BlueScore          uint64
	ChildrenHashes     []string
}
