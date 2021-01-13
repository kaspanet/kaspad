package appmessage

// MsgIBDRootUTXOSetChunk represents a kaspa IBDRootUTXOSetChunk message
type MsgIBDRootUTXOSetChunk struct {
	baseMessage
	Chunk []byte
}

// Command returns the protocol command string for the message
func (msg *MsgIBDRootUTXOSetChunk) Command() MessageCommand {
	return CmdIBDRootUTXOSetChunk
}

// NewMsgIBDRootUTXOSetChunk returns a new MsgIBDRootUTXOSetChunk.
func NewMsgIBDRootUTXOSetChunk(chunk []byte) *MsgIBDRootUTXOSetChunk {
	return &MsgIBDRootUTXOSetChunk{
		Chunk: chunk,
	}
}
