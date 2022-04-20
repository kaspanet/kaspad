package appmessage

// GetBlockTemplateRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateRequestMessage struct {
	baseMessage
	PayAddress string
	ExtraData  string
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateRequestMessage) Command() MessageCommand {
	return CmdGetBlockTemplateRequestMessage
}

// NewGetBlockTemplateRequestMessage returns a instance of the message
func NewGetBlockTemplateRequestMessage(payAddress, extraData string) *GetBlockTemplateRequestMessage {
	return &GetBlockTemplateRequestMessage{
		PayAddress: payAddress,
		ExtraData:  extraData,
	}
}

// GetBlockTemplateResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockTemplateResponseMessage struct {
	baseMessage
	Block     *RPCBlock
	IsSynced  bool
	Donations []*Donation

	Error *RPCError
}

// Donation is an appmessage corresponding to
// its respective RPC message
type Donation struct {
	DonationAddress string
	DonationPercent float32
}

// Command returns the protocol command string for the message
func (msg *GetBlockTemplateResponseMessage) Command() MessageCommand {
	return CmdGetBlockTemplateResponseMessage
}

// NewGetBlockTemplateResponseMessage returns a instance of the message
func NewGetBlockTemplateResponseMessage(block *RPCBlock, isSynced bool, donations []*Donation) *GetBlockTemplateResponseMessage {
	return &GetBlockTemplateResponseMessage{
		Block:     block,
		IsSynced:  isSynced,
		Donations: donations,
	}
}
