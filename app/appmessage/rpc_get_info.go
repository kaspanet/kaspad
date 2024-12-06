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
	P2PID                     string
	MempoolSize               uint64
	ServerVersion             string
	IsUtxoIndexed             bool
	IsSynced                  bool
	MaxRPCClients             int64
	NumberOfRPCConnections    int64
	MaxP2PClients             int64
	NumberOfP2PConnections    int64
	BanDurationInMilliseconds int64
	UptimeInMilliseconds      int64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetInfoResponseMessage) Command() MessageCommand {
	return CmdGetInfoResponseMessage
}

// NewGetInfoResponseMessage returns a instance of the message
func NewGetInfoResponseMessage(p2pID string, mempoolSize uint64, serverVersion string,
	isUtxoIndexed bool, isSynced bool, maxRPCClients int64, numberOfRPCConnections int64,
	maxP2PClients int64, numberOfP2PConnections int64, banDurationInMilliseconds int64,
	uptimeInMilliseconds int64) *GetInfoResponseMessage {
	return &GetInfoResponseMessage{
		P2PID:                     p2pID,
		MempoolSize:               mempoolSize,
		ServerVersion:             serverVersion,
		IsUtxoIndexed:             isUtxoIndexed,
		IsSynced:                  isSynced,
		MaxRPCClients:             maxRPCClients,
		NumberOfRPCConnections:    numberOfRPCConnections,
		MaxP2PClients:             maxP2PClients,
		NumberOfP2PConnections:    numberOfP2PConnections,
		BanDurationInMilliseconds: banDurationInMilliseconds,
		UptimeInMilliseconds:      uptimeInMilliseconds,
	}
}
