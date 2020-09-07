package appmessage

// GetSelectedTipHashRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetSelectedTipHashRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetSelectedTipHashRequestMessage) Command() MessageCommand {
	return CmdGetSelectedTipHashRequestMessage
}

// NewGetSelectedTipHashRequestMessage returns a instance of the message
func NewGetSelectedTipHashRequestMessage() *GetSelectedTipHashRequestMessage {
	return &GetSelectedTipHashRequestMessage{}
}

// GetSelectedTipHashResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetSelectedTipHashResponseMessage struct {
	baseMessage
	SelectedTipHash string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetSelectedTipHashResponseMessage) Command() MessageCommand {
	return CmdGetSelectedTipHashResponseMessage
}

// NewGetSelectedTipHashResponseMessage returns a instance of the message
func NewGetSelectedTipHashResponseMessage(selectedTipHash string) *GetSelectedTipHashResponseMessage {
	return &GetSelectedTipHashResponseMessage{
		SelectedTipHash: selectedTipHash,
	}
}
