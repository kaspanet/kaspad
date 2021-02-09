package appmessage

// SubmitBlockRequestMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockRequestMessage struct {
	baseMessage
	Block *MsgBlock
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockRequestMessage) Command() MessageCommand {
	return CmdSubmitBlockRequestMessage
}

// NewSubmitBlockRequestMessage returns a instance of the message
func NewSubmitBlockRequestMessage(block *MsgBlock) *SubmitBlockRequestMessage {
	return &SubmitBlockRequestMessage{
		Block: block,
	}
}

// RejectReason describes the reason why a block sent by SubmitBlock was rejected
type RejectReason byte

// RejectReason constants
// Not using iota, since in the .proto file those are hardcoded
const (
	RejectReasonNone         RejectReason = 0
	RejectReasonBlockInvalid RejectReason = 1
	RejectReasonIsInIBD      RejectReason = 2
)

var rejectReasonToString = map[RejectReason]string{
	RejectReasonNone:         "None",
	RejectReasonBlockInvalid: "Block is invalid",
	RejectReasonIsInIBD:      "Node is in IBD",
}

func (rr RejectReason) String() string {
	return rejectReasonToString[rr]
}

// SubmitBlockResponseMessage is an appmessage corresponding to
// its respective RPC message
type SubmitBlockResponseMessage struct {
	baseMessage
	RejectReason RejectReason
	Error        *RPCError
}

// Command returns the protocol command string for the message
func (msg *SubmitBlockResponseMessage) Command() MessageCommand {
	return CmdSubmitBlockResponseMessage
}

// NewSubmitBlockResponseMessage returns a instance of the message
func NewSubmitBlockResponseMessage() *SubmitBlockResponseMessage {
	return &SubmitBlockResponseMessage{}
}
