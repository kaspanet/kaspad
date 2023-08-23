package appmessage

// MsgRequestNextPruningPointUTXOSetChunk represents a c4ex RequestNextPruningPointUTXOSetChunk message
type MsgRequestNextPruningPointUTXOSetChunk struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgRequestNextPruningPointUTXOSetChunk) Command() MessageCommand {
	return CmdRequestNextPruningPointUTXOSetChunk
}

// NewMsgRequestNextPruningPointUTXOSetChunk returns a new MsgRequestNextPruningPointUTXOSetChunk.
func NewMsgRequestNextPruningPointUTXOSetChunk() *MsgRequestNextPruningPointUTXOSetChunk {
	return &MsgRequestNextPruningPointUTXOSetChunk{}
}
