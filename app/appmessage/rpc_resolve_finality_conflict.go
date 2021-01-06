package appmessage

// ResolveFinalityConflictRequestMessage is an appmessage corresponding to
// its respective RPC message
type ResolveFinalityConflictRequestMessage struct {
	baseMessage
	FinalityBlockHash string
}

// Command returns the protocol command string for the message
func (msg *ResolveFinalityConflictRequestMessage) Command() MessageCommand {
	return CmdResolveFinalityConflictRequestMessage
}

// NewResolveFinalityConflictRequestMessage returns a instance of the message
func NewResolveFinalityConflictRequestMessage(finalityBlockHash string) *ResolveFinalityConflictRequestMessage {
	return &ResolveFinalityConflictRequestMessage{
		FinalityBlockHash: finalityBlockHash,
	}
}

// ResolveFinalityConflictResponseMessage is an appmessage corresponding to
// its respective RPC message
type ResolveFinalityConflictResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *ResolveFinalityConflictResponseMessage) Command() MessageCommand {
	return CmdResolveFinalityConflictResponseMessage
}

// NewResolveFinalityConflictResponseMessage returns a instance of the message
func NewResolveFinalityConflictResponseMessage() *ResolveFinalityConflictResponseMessage {
	return &ResolveFinalityConflictResponseMessage{}
}
