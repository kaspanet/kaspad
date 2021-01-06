package appmessage

// GetVirtualSelectedParentBlueScoreRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetVirtualSelectedParentBlueScoreRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetVirtualSelectedParentBlueScoreRequestMessage) Command() MessageCommand {
	return CmdGetVirtualSelectedParentBlueScoreRequestMessage
}

// NewGetVirtualSelectedParentBlueScoreRequestMessage returns a instance of the message
func NewGetVirtualSelectedParentBlueScoreRequestMessage() *GetVirtualSelectedParentBlueScoreRequestMessage {
	return &GetVirtualSelectedParentBlueScoreRequestMessage{}
}

// GetVirtualSelectedParentBlueScoreResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetVirtualSelectedParentBlueScoreResponseMessage struct {
	baseMessage
	BlueScore uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetVirtualSelectedParentBlueScoreResponseMessage) Command() MessageCommand {
	return CmdGetVirtualSelectedParentBlueScoreResponseMessage
}

// NewGetVirtualSelectedParentBlueScoreResponseMessage returns a instance of the message
func NewGetVirtualSelectedParentBlueScoreResponseMessage(blueScore uint64) *GetVirtualSelectedParentBlueScoreResponseMessage {
	return &GetVirtualSelectedParentBlueScoreResponseMessage{
		BlueScore: blueScore,
	}
}
