package appmessage

// GetBlockHexRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockHexRequestMessage struct {
	baseMessage
	Hash         string
	SubnetworkID string
}

// Command returns the protocol command string for the message
func (msg *GetBlockHexRequestMessage) Command() MessageCommand {
	return CmdGetBlockHexRequestMessage
}

// GetBlockHexRequestMessage returns a instance of the message
func NewGetBlockHexRequestMessage(hash string, subnetworkID string) *GetBlockHexRequestMessage {
	return &GetBlockHexRequestMessage{
		Hash:         hash,
		SubnetworkID: subnetworkID,
	}
}

// GetBlockHexResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockHexResponseMessage struct {
	baseMessage
	BlockHex string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockHexResponseMessage) Command() MessageCommand {
	return CmdGetBlockHexResponseMessage
}

// GetBlockHexResponseMessage returns a instance of the message
func NewGetBlockHexResponseMessage(blockHex string) *GetBlockHexResponseMessage {
	return &GetBlockHexResponseMessage{
		BlockHex: blockHex,
	}
}
