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
	MempoolSize   uint64
	ServerVersion string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetInfoResponseMessage) Command() MessageCommand {
	return CmdGetInfoResponseMessage
}

// NewGetInfoResponseMessage returns a instance of the message
func NewGetInfoResponseMessage(p2pID string, mempoolSize uint64, serverVersion string) *GetInfoResponseMessage {
	return &GetInfoResponseMessage{
		P2PID:         p2pID,
		MempoolSize:   mempoolSize,
		ServerVersion: serverVersion,
	}
}
