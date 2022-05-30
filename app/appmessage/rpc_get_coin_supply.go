package appmessage

// GetInfoRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetCoinSupplyRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetCoinSupplyRequestMessage) Command() MessageCommand {
	return CmdGetCoinSupplyRequestMessage
}

// NewGetInfoRequestMessage returns a instance of the message
func NewGetCoinSupplyRequestMessage() *GetCoinSupplyRequestMessage {
	return &GetCoinSupplyRequestMessage{}
}

// GetInfoResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetCoinSupplyResponseMessage struct {
	baseMessage
	TotalSompi uint64
	CirculatingSompi uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetCoinSupplyResponseMessage) Command() MessageCommand {
	return CmdGetCoinSupplyResponseMessage
}

// NewGetInfoResponseMessage returns a instance of the message
func NewGetCoinSupplyResponseMessage(totalSompi uint64, circulatingSompi uint64) *GetCoinSupplyResponseMessage {
	return &GetCoinSupplyResponseMessage{
		TotalSompi:    		totalSompi,
		CirculatingSompi:	circulatingSompi,
	}
}
