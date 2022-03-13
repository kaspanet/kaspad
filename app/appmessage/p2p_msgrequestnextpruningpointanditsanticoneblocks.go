package appmessage

// MsgRequestNextPruningPointAndItsAnticoneBlocks implements the Message interface and represents a kaspa
// RequestNextPruningPointAndItsAnticoneBlocks message. It is used to notify the IBD syncer peer to send
// more blocks from the pruning anticone.
//
// This message has no payload.
type MsgRequestNextPruningPointAndItsAnticoneBlocks struct {
	baseMessage
}

// Command returns the protocol command string for the message. This is part
// of the Message interface implementation.
func (msg *MsgRequestNextPruningPointAndItsAnticoneBlocks) Command() MessageCommand {
	return CmdRequestNextPruningPointAndItsAnticoneBlocks
}

// NewMsgRequestNextPruningPointAndItsAnticoneBlocks returns a new kaspa RequestNextPruningPointAndItsAnticoneBlocks message that conforms to the
// Message interface.
func NewMsgRequestNextPruningPointAndItsAnticoneBlocks() *MsgRequestNextPruningPointAndItsAnticoneBlocks {
	return &MsgRequestNextPruningPointAndItsAnticoneBlocks{}
}
