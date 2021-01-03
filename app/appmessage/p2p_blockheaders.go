package appmessage

// BlockHeadersMessage represents a kaspa BlockHeaders message
type BlockHeadersMessage struct {
	baseMessage
	BlockHeaders []*MsgBlockHeader
}

// Command returns the protocol command string for the message
func (msg *BlockHeadersMessage) Command() MessageCommand {
	return CmdBlockHeaders
}

// NewMsgBlockHeadersMessage returns a new kaspa BlockHeaders message
func NewMsgBlockHeadersMessage(blockHeaders []*MsgBlockHeader) *BlockHeadersMessage {
	return &BlockHeadersMessage{
		BlockHeaders: blockHeaders,
	}
}
