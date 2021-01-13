package appmessage

// MsgRequestNextIBDRootUTXOSetChunk represents a kaspa RequestNextIBDRootUTXOSetChunk message
type MsgRequestNextIBDRootUTXOSetChunk struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgRequestNextIBDRootUTXOSetChunk) Command() MessageCommand {
	return CmdRequestNextIBDRootUTXOSetChunk
}

// NewMsgRequestNextIBDRootUTXOSetChunk returns a new MsgRequestNextIBDRootUTXOSetChunk.
func NewMsgRequestNextIBDRootUTXOSetChunk(chunk []byte) *MsgRequestNextIBDRootUTXOSetChunk {
	return &MsgRequestNextIBDRootUTXOSetChunk{}
}
