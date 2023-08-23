package appmessage

// MsgRequestPruningPointProof represents a c4ex RequestPruningPointProof message
type MsgRequestPruningPointProof struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgRequestPruningPointProof) Command() MessageCommand {
	return CmdRequestPruningPointProof
}

// NewMsgRequestPruningPointProof returns a new MsgRequestPruningPointProof.
func NewMsgRequestPruningPointProof() *MsgRequestPruningPointProof {
	return &MsgRequestPruningPointProof{}
}
