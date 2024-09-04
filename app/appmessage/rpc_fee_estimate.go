package appmessage

// GetFeeEstimateRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetFeeEstimateRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetFeeEstimateRequestMessage) Command() MessageCommand {
	return CmdGetFeeEstimateRequestMessage
}

// NewGetFeeEstimateRequestMessage returns a instance of the message
func NewGetFeeEstimateRequestMessage() *GetFeeEstimateRequestMessage {
	return &GetFeeEstimateRequestMessage{}
}

type RPCFeeRateBucket struct {
	Feerate          float64
	EstimatedSeconds float64
}

type RPCFeeEstimate struct {
	PriorityBucket RPCFeeRateBucket
	NormalBuckets  []RPCFeeRateBucket
	LowBuckets     []RPCFeeRateBucket
}

// GetCoinSupplyResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetFeeEstimateResponseMessage struct {
	baseMessage
	Estimate RPCFeeEstimate

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetFeeEstimateResponseMessage) Command() MessageCommand {
	return CmdGetFeeEstimateResponseMessage
}

// NewGetFeeEstimateResponseMessage returns a instance of the message
func NewGetFeeEstimateResponseMessage() *GetFeeEstimateResponseMessage {
	return &GetFeeEstimateResponseMessage{}
}
