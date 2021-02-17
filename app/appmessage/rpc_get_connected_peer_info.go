package appmessage

// GetConnectedPeerInfoRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetConnectedPeerInfoRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetConnectedPeerInfoRequestMessage) Command() MessageCommand {
	return CmdGetConnectedPeerInfoRequestMessage
}

// NewGetConnectedPeerInfoRequestMessage returns a instance of the message
func NewGetConnectedPeerInfoRequestMessage() *GetConnectedPeerInfoRequestMessage {
	return &GetConnectedPeerInfoRequestMessage{}
}

// GetConnectedPeerInfoResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetConnectedPeerInfoResponseMessage struct {
	baseMessage
	Infos []*GetConnectedPeerInfoMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetConnectedPeerInfoResponseMessage) Command() MessageCommand {
	return CmdGetConnectedPeerInfoResponseMessage
}

// NewGetConnectedPeerInfoResponseMessage returns a instance of the message
func NewGetConnectedPeerInfoResponseMessage(infos []*GetConnectedPeerInfoMessage) *GetConnectedPeerInfoResponseMessage {
	return &GetConnectedPeerInfoResponseMessage{
		Infos: infos,
	}
}

// GetConnectedPeerInfoMessage holds information about a connected peer
type GetConnectedPeerInfoMessage struct {
	ID                        string
	Address                   string
	LastPingDuration          int64
	IsOutbound                bool
	TimeOffset                int64
	UserAgent                 string
	AdvertisedProtocolVersion uint32
	TimeConnected             int64
	IsIBDPeer                 bool
}
