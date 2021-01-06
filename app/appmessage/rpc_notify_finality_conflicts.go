package appmessage

// NotifyFinalityConflictsRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyFinalityConflictsRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *NotifyFinalityConflictsRequestMessage) Command() MessageCommand {
	return CmdNotifyFinalityConflictsRequestMessage
}

// NewNotifyFinalityConflictsRequestMessage returns a instance of the message
func NewNotifyFinalityConflictsRequestMessage() *NotifyFinalityConflictsRequestMessage {
	return &NotifyFinalityConflictsRequestMessage{}
}

// NotifyFinalityConflictsResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyFinalityConflictsResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyFinalityConflictsResponseMessage) Command() MessageCommand {
	return CmdNotifyFinalityConflictsResponseMessage
}

// NewNotifyFinalityConflictsResponseMessage returns a instance of the message
func NewNotifyFinalityConflictsResponseMessage() *NotifyFinalityConflictsResponseMessage {
	return &NotifyFinalityConflictsResponseMessage{}
}

// FinalityConflictNotificationMessage is an appmessage corresponding to
// its respective RPC message
type FinalityConflictNotificationMessage struct {
	baseMessage
	ViolatingBlockHash string
}

// Command returns the protocol command string for the message
func (msg *FinalityConflictNotificationMessage) Command() MessageCommand {
	return CmdFinalityConflictNotificationMessage
}

// NewFinalityConflictNotificationMessage returns a instance of the message
func NewFinalityConflictNotificationMessage(violatingBlockHash string) *FinalityConflictNotificationMessage {
	return &FinalityConflictNotificationMessage{
		ViolatingBlockHash: violatingBlockHash,
	}
}

// FinalityConflictResolvedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type FinalityConflictResolvedNotificationMessage struct {
	baseMessage
	FinalityBlockHash string
}

// Command returns the protocol command string for the message
func (msg *FinalityConflictResolvedNotificationMessage) Command() MessageCommand {
	return CmdFinalityConflictResolvedNotificationMessage
}

// NewFinalityConflictResolvedNotificationMessage returns a instance of the message
func NewFinalityConflictResolvedNotificationMessage(finalityBlockHash string) *FinalityConflictResolvedNotificationMessage {
	return &FinalityConflictResolvedNotificationMessage{
		FinalityBlockHash: finalityBlockHash,
	}
}
