package appmessage

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
