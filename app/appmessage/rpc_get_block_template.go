package appmessage

// GetBlockTemplateRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateRequestMessage struct {
	baseMessage
	PayAddress string
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateRequestMessage) Command() MessageCommand {
	return CmdGetBlockTemplateRequestMessage
}

// NewGetBlockTemplateRequestMessage returns a instance of the message
func NewGetBlockTemplateRequestMessage(payAddress string) *GetBlockTemplateRequestMessage {
	return &GetBlockTemplateRequestMessage{
		PayAddress: payAddress,
	}
}

// GetBlockTemplateResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateResponseMessage struct {
	baseMessage
	Block    *RPCBlock
	IsSynced bool

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateResponseMessage) Command() MessageCommand {
	return CmdGetBlockTemplateResponseMessage
}

// NewGetBlockTemplateResponseMessage returns a instance of the message
func NewGetBlockTemplateResponseMessage(block *RPCBlock, isSynced bool) *GetBlockTemplateResponseMessage {
	return &GetBlockTemplateResponseMessage{
		Block:    block,
		IsSynced: isSynced,
	}
}
