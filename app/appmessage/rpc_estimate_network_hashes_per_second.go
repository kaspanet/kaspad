package appmessage

// EstimateNetworkHashesPerSecondRequestMessage is an appmessage corresponding to
// its respective RPC message
type EstimateNetworkHashesPerSecondRequestMessage struct {
	baseMessage
	StartHash  string
	WindowSize uint32
}

// Command returns the protocol command string for the message
func (msg *EstimateNetworkHashesPerSecondRequestMessage) Command() MessageCommand {
	return CmdEstimateNetworkHashesPerSecondRequestMessage
}

// NewEstimateNetworkHashesPerSecondRequestMessage returns a instance of the message
func NewEstimateNetworkHashesPerSecondRequestMessage(startHash string, windowSize uint32) *EstimateNetworkHashesPerSecondRequestMessage {
	return &EstimateNetworkHashesPerSecondRequestMessage{
		StartHash:  startHash,
		WindowSize: windowSize,
	}
}

// EstimateNetworkHashesPerSecondResponseMessage is an appmessage corresponding to
// its respective RPC message
type EstimateNetworkHashesPerSecondResponseMessage struct {
	baseMessage
	NetworkHashesPerSecond uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *EstimateNetworkHashesPerSecondResponseMessage) Command() MessageCommand {
	return CmdEstimateNetworkHashesPerSecondResponseMessage
}

// NewEstimateNetworkHashesPerSecondResponseMessage returns a instance of the message
func NewEstimateNetworkHashesPerSecondResponseMessage(networkHashesPerSecond uint64) *EstimateNetworkHashesPerSecondResponseMessage {
	return &EstimateNetworkHashesPerSecondResponseMessage{
		NetworkHashesPerSecond: networkHashesPerSecond,
	}
}
