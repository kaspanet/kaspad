package appmessage

// MsgPruningPointProof represents a kaspa PruningPointProof message
type MsgPruningPointProof struct {
	baseMessage

	Headers []*MsgBlockHeader
}

// Command returns the protocol command string for the message
func (msg *MsgPruningPointProof) Command() MessageCommand {
	return CmdPruningPointProof
}

// NewMsgPruningPointProof returns a new MsgPruningPointProof.
func NewMsgPruningPointProof(headers []*MsgBlockHeader) *MsgPruningPointProof {
	return &MsgPruningPointProof{
		Headers: headers,
	}
}
