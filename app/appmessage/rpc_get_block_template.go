package appmessage

// GetBlockTemplateRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateRequestMessage struct {
	baseMessage
	PayAddress string
	LongPollID string
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateRequestMessage) Command() MessageCommand {
	return CmdGetBlockTemplateRequestMessage
}

// GetBlockTemplateRequestMessage returns a instance of the message
func NewGetBlockTemplateRequestMessage(payAddress string, longPollID string) *GetBlockTemplateRequestMessage {
	return &GetBlockTemplateRequestMessage{
		PayAddress: payAddress,
		LongPollID: longPollID,
	}
}

// GetBlockTemplateResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateResponseMessage struct {
	baseMessage
	Bits                 string
	CurrentTime          int64
	ParentHashes         []string
	MassLimit            int
	Transactions         []GetBlockTemplateTransactionMessage
	HashMerkleRoot       string
	AcceptedIDMerkleRoot string
	UTXOCommitment       string
	Version              int32
	LongPollID           string
	TargetDifficulty     string
	MinTime              int64
	MaxTime              int64
	MutableFields        []string
	NonceRange           string
	IsSynced             bool
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateResponseMessage) Command() MessageCommand {
	return CmdGetBlockTemplateResponseMessage
}

// GetBlockTemplateResponseMessage returns a instance of the message
func NewGetBlockTemplateResponseMessage() *GetBlockTemplateResponseMessage {
	return &GetBlockTemplateResponseMessage{}
}

// GetBlockTemplateTransactionMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateTransactionMessage struct {
	baseMessage
	Data    string
	ID      string
	Depends []int64
	Mass    uint64
	Fee     uint64
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateTransactionMessage) Command() MessageCommand {
	return CmdGetBlockTemplateTransactionMessage
}

// GetBlockTemplateTransactionMessage returns a instance of the message
func NewGetBlockTemplateTransactionMessage() *GetBlockTemplateTransactionMessage {
	return &GetBlockTemplateTransactionMessage{}
}
