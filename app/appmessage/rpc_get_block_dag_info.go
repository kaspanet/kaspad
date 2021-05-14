package appmessage

// GetBlockDAGInfoRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockDAGInfoRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetBlockDAGInfoRequestMessage) Command() MessageCommand {
	return CmdGetBlockDAGInfoRequestMessage
}

// NewGetBlockDAGInfoRequestMessage returns a instance of the message
func NewGetBlockDAGInfoRequestMessage() *GetBlockDAGInfoRequestMessage {
	return &GetBlockDAGInfoRequestMessage{}
}

// GetBlockDAGInfoResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockDAGInfoResponseMessage struct {
	baseMessage
	NetworkName         string
	BlockCount          uint64
	HeaderCount         uint64
	TipHashes           []string
	VirtualParentHashes []string
	Difficulty          float64
	PastMedianTime      int64
	PruningPointHash    string
	VirtualDAAScore     uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockDAGInfoResponseMessage) Command() MessageCommand {
	return CmdGetBlockDAGInfoResponseMessage
}

// NewGetBlockDAGInfoResponseMessage returns a instance of the message
func NewGetBlockDAGInfoResponseMessage() *GetBlockDAGInfoResponseMessage {
	return &GetBlockDAGInfoResponseMessage{}
}
