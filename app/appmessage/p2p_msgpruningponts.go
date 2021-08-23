package appmessage

// MsgPruningPoints represents a kaspa PruningPoints message
type MsgPruningPoints struct {
	baseMessage

	Headers []*MsgBlockHeader
}

// Command returns the protocol command string for the message
func (msg *MsgPruningPoints) Command() MessageCommand {
	return CmdPruningPoints
}

// NewMsgPruningPoints returns a new MsgPruningPoints.
func NewMsgPruningPoints(headers []*MsgBlockHeader) *MsgPruningPoints {
	return &MsgPruningPoints{
		Headers: headers,
	}
}
