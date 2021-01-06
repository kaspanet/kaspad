package appmessage

// GetSubnetworkRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetSubnetworkRequestMessage struct {
	baseMessage
	SubnetworkID string
}

// Command returns the protocol command string for the message
func (msg *GetSubnetworkRequestMessage) Command() MessageCommand {
	return CmdGetSubnetworkRequestMessage
}

// NewGetSubnetworkRequestMessage returns a instance of the message
func NewGetSubnetworkRequestMessage(subnetworkID string) *GetSubnetworkRequestMessage {
	return &GetSubnetworkRequestMessage{
		SubnetworkID: subnetworkID,
	}
}

// GetSubnetworkResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetSubnetworkResponseMessage struct {
	baseMessage
	GasLimit uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetSubnetworkResponseMessage) Command() MessageCommand {
	return CmdGetSubnetworkResponseMessage
}

// NewGetSubnetworkResponseMessage returns a instance of the message
func NewGetSubnetworkResponseMessage(gasLimit uint64) *GetSubnetworkResponseMessage {
	return &GetSubnetworkResponseMessage{
		GasLimit: gasLimit,
	}
}
