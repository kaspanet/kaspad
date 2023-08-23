package appmessage

// BlockHeadersMessage represents a c4ex BlockHeaders message
type BlockHeadersMessage struct {
	baseMessage
	BlockHeaders []*MsgBlockHeader
}

// Command returns the protocol command string for the message
func (msg *BlockHeadersMessage) Command() MessageCommand {
	return CmdBlockHeaders
}

// NewBlockHeadersMessage returns a new c4ex BlockHeaders message
func NewBlockHeadersMessage(blockHeaders []*MsgBlockHeader) *BlockHeadersMessage {
	return &BlockHeadersMessage{
		BlockHeaders: blockHeaders,
	}
}
