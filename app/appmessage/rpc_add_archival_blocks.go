package appmessage

type ArchivalBlock struct {
	Block *RPCBlock
	Child string
}

// AddArchivalBlocksRequestMessage represents a request to add archival blocks
type AddArchivalBlocksRequestMessage struct {
	baseMessage
	Blocks []*ArchivalBlock
}

// Command returns the protocol command string for the message
func (msg *AddArchivalBlocksRequestMessage) Command() MessageCommand {
	return CmdAddArchivalBlocksRequestMessage
}

// NewAddArchivalBlocksRequestMessage returns a instance of the message
func NewAddArchivalBlocksRequestMessage(blocks []*ArchivalBlock) *AddArchivalBlocksRequestMessage {
	return &AddArchivalBlocksRequestMessage{
		Blocks: blocks,
	}
}

// AddArchivalBlocksResponseMessage represents a response to the AddArchivalBlocks request
type AddArchivalBlocksResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *AddArchivalBlocksResponseMessage) Command() MessageCommand {
	return CmdAddArchivalBlocksResponseMessage
}

// NewAddArchivalBlocksResponseMessage returns a instance of the message
func NewAddArchivalBlocksResponseMessage(err *RPCError) *AddArchivalBlocksResponseMessage {
	return &AddArchivalBlocksResponseMessage{
		Error: err,
	}
}
