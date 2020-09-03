package appmessage

// GetBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockRequestMessage struct {
	baseMessage
	Hash                    string
	SubnetworkID            string
	IncludeBlockHex         bool
	IncludeBlockVerboseData bool
}

// Command returns the protocol command string for the message
func (msg *GetBlockRequestMessage) Command() MessageCommand {
	return CmdGetBlockRequestMessage
}

// GetBlockRequestMessage returns a instance of the message
func NewGetBlockRequestMessage(hash string, subnetworkID string, includeBlockHex bool, includeBlockVerboseData bool) *GetBlockRequestMessage {
	return &GetBlockRequestMessage{
		Hash:                    hash,
		SubnetworkID:            subnetworkID,
		IncludeBlockHex:         includeBlockHex,
		IncludeBlockVerboseData: includeBlockVerboseData,
	}
}

// GetBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockResponseMessage struct {
	baseMessage
	BlockHex string

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
