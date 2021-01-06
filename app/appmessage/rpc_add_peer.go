package appmessage

// AddPeerRequestMessage is an appmessage corresponding to
// its respective RPC message
type AddPeerRequestMessage struct {
	baseMessage
	Address     string
	IsPermanent bool
}

// Command returns the protocol command string for the message
func (msg *AddPeerRequestMessage) Command() MessageCommand {
	return CmdAddPeerRequestMessage
}

// NewAddPeerRequestMessage returns a instance of the message
func NewAddPeerRequestMessage(address string, isPermanent bool) *AddPeerRequestMessage {
	return &AddPeerRequestMessage{
		Address:     address,
		IsPermanent: isPermanent,
	}
}

// AddPeerResponseMessage is an appmessage corresponding to
// its respective RPC message
type AddPeerResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *AddPeerResponseMessage) Command() MessageCommand {
	return CmdAddPeerResponseMessage
}

// NewAddPeerResponseMessage returns a instance of the message
func NewAddPeerResponseMessage() *AddPeerResponseMessage {
	return &AddPeerResponseMessage{}
}
