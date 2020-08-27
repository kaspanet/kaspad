package appmessage

// GetBlockTemplateRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateRequestMessage) Command() MessageCommand {
	return CmdGetBlockTemplateRequestMessage
}

// GetBlockTemplateRequestMessage returns a instance of the message
func NewGetBlockTemplateRequestMessage() *GetBlockTemplateRequestMessage {
	return &GetBlockTemplateRequestMessage{}
}
