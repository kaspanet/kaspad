package appmessage

// GetInfoRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetInfoRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetInfoRequestMessage) Command() MessageCommand {
	return CmdGetInfoRequestMessage
}

// NewGetInfoRequestMessage returns a instance of the message
func NewGetInfoRequestMessage() *GetInfoRequestMessage {
	return &GetInfoRequestMessage{}
}

// GetInfoResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetInfoResponseMessage struct {
	baseMessage
	P2PID         string
	ServerVersion string
	MempoolSize   uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetInfoResponseMessage) Command() MessageCommand {
	return CmdGetInfoResponseMessage
}

// NewGetInfoResponseMessage returns a instance of the message
func NewGetInfoResponseMessage(p2pID string, serverVersion string, mempoolSize uint64) *GetInfoResponseMessage {
	return &GetInfoResponseMessage{
		P2PID:         p2pID,
		ServerVersion: serverVersion,
		MempoolSize:   mempoolSize,
	}
}
