package appmessage

// IBDBlocksMessage represents a kaspa IBDBlocks message
type IBDBlocksMessage struct {
	baseMessage
	Blocks []*MsgBlock
}

// Command returns the protocol command string for the message
func (msg *IBDBlocksMessage) Command() MessageCommand {
	return CmdIBDBlocks
}

// NewBlockHeadersMessage returns a new kaspa BlockHeaders message
func NewBlockHeadersMessage(blocks []*MsgBlock) *IBDBlocksMessage {
	return &IBDBlocksMessage{
		Blocks: blocks,
	}
}
