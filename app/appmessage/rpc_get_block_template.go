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

// GetBlockTemplateRequestMessage returns a instance of the message
func NewGetBlockTemplateRequestMessage() *GetBlockTemplateRequestMessage {
	return &GetBlockTemplateRequestMessage{}
}

// GetBlockTemplateResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateResponseMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateResponseMessage) Command() MessageCommand {
	return CmdGetBlockTemplateResponseMessage
}

// GetBlockTemplateResponseMessage returns a instance of the message
func NewGetBlockTemplateResponseMessage() *GetBlockTemplateResponseMessage {
	return &GetBlockTemplateResponseMessage{}
}
