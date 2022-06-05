package appmessage

// GetCoinSupplyRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetCoinSupplyRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetCoinSupplyRequestMessage) Command() MessageCommand {
	return CmdGetCoinSupplyRequestMessage
}

// NewGetCoinSupplyRequestMessage returns a instance of the message
func NewGetCoinSupplyRequestMessage() *GetCoinSupplyRequestMessage {
	return &GetCoinSupplyRequestMessage{}
}

// GetCoinSupplyResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetCoinSupplyResponseMessage struct {
	baseMessage
	MaxSompi         uint64
	CirculatingSompi uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetCoinSupplyResponseMessage) Command() MessageCommand {
	return CmdGetCoinSupplyResponseMessage
}

// NewGetCoinSupplyResponseMessage returns a instance of the message
func NewGetCoinSupplyResponseMessage(maxSompi uint64, circulatingSompi uint64) *GetCoinSupplyResponseMessage {
	return &GetCoinSupplyResponseMessage{
		MaxSompi:         maxSompi,
		CirculatingSompi: circulatingSompi,
	}
}
