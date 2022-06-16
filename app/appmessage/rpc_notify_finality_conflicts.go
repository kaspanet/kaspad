package appmessage

// NotifyFinalityConflictsRequestMessage is an appmessage corresponding to
// its respective RPC message
type NotifyFinalityConflictsRequestMessage struct {
	baseMessage
	Id string
}

// Command returns the protocol command string for the message
func (msg *NotifyFinalityConflictsRequestMessage) Command() MessageCommand {
	return CmdNotifyFinalityConflictsRequestMessage
}

// NewNotifyFinalityConflictsRequestMessage returns a instance of the message
func NewNotifyFinalityConflictsRequestMessage(id string) *NotifyFinalityConflictsRequestMessage {
	return &NotifyFinalityConflictsRequestMessage{Id: id}
}

// NotifyFinalityConflictsResponseMessage is an appmessage corresponding to
// its respective RPC message
type NotifyFinalityConflictsResponseMessage struct {
	baseMessage
	Id    string
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *NotifyFinalityConflictsResponseMessage) Command() MessageCommand {
	return CmdNotifyFinalityConflictsResponseMessage
}

// NewNotifyFinalityConflictsResponseMessage returns a instance of the message
func NewNotifyFinalityConflictsResponseMessage(id string) *NotifyFinalityConflictsResponseMessage {
	return &NotifyFinalityConflictsResponseMessage{Id: id}
}

// FinalityConflictNotificationMessage is an appmessage corresponding to
// its respective RPC message
type FinalityConflictNotificationMessage struct {
	baseMessage
	Id                 string
	ViolatingBlockHash string
}

// Command returns the protocol command string for the message
func (msg *FinalityConflictNotificationMessage) Command() MessageCommand {
	return CmdFinalityConflictNotificationMessage
}

// NewFinalityConflictNotificationMessage returns a instance of the message
func NewFinalityConflictNotificationMessage(violatingBlockHash string, id string) *FinalityConflictNotificationMessage {
	return &FinalityConflictNotificationMessage{
		ViolatingBlockHash: violatingBlockHash,
	}
}

// FinalityConflictResolvedNotificationMessage is an appmessage corresponding to
// its respective RPC message
type FinalityConflictResolvedNotificationMessage struct {
	baseMessage
	Id                string
	FinalityBlockHash string
}

// Command returns the protocol command string for the message
func (msg *FinalityConflictResolvedNotificationMessage) Command() MessageCommand {
	return CmdFinalityConflictResolvedNotificationMessage
}

// NewFinalityConflictResolvedNotificationMessage returns a instance of the message
func NewFinalityConflictResolvedNotificationMessage(finalityBlockHash string, id string) *FinalityConflictResolvedNotificationMessage {
	return &FinalityConflictResolvedNotificationMessage{
		Id:                id,
		FinalityBlockHash: finalityBlockHash,
	}
}
