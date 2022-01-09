package appmessage

// MsgBlockWithTrustedDataV4 represents a kaspa BlockWithTrustedDataV4 message
type MsgBlockWithTrustedDataV4 struct {
	baseMessage

	Block               *MsgBlock
	DAAWindowIndices    []uint64
	GHOSTDAGDataIndices []uint64
}

// Command returns the protocol command string for the message
func (msg *MsgBlockWithTrustedDataV4) Command() MessageCommand {
	return CmdBlockWithTrustedDataV4
}

// NewMsgBlockWithTrustedDataV4 returns a new MsgBlockWithTrustedDataV4.
func NewMsgBlockWithTrustedDataV4() *MsgBlockWithTrustedDataV4 {
	return &MsgBlockWithTrustedDataV4{}
}
