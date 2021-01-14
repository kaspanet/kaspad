package appmessage

// MsgDoneIBDRootUTXOSetChunks represents a kaspa DoneIBDRootUTXOSetChunks message
type MsgDoneIBDRootUTXOSetChunks struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *MsgDoneIBDRootUTXOSetChunks) Command() MessageCommand {
	return CmdDoneIBDRootUTXOSetChunks
}

// NewMsgDoneIBDRootUTXOSetChunks returns a new MsgDoneIBDRootUTXOSetChunks.
func NewMsgDoneIBDRootUTXOSetChunks() *MsgDoneIBDRootUTXOSetChunks {
	return &MsgDoneIBDRootUTXOSetChunks{}
}
