package appmessage

// GetHeadersRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetHeadersRequestMessage struct {
	baseMessage
	StartHash   string
	Limit       uint64
	IsAscending bool
}

// Command returns the protocol command string for the message
func (msg *GetHeadersRequestMessage) Command() MessageCommand {
	return CmdGetHeadersRequestMessage
}

// NewGetHeadersRequestMessage returns a instance of the message
func NewGetHeadersRequestMessage(startHash string, limit uint64, isAscending bool) *GetHeadersRequestMessage {
	return &GetHeadersRequestMessage{
		StartHash:   startHash,
		Limit:       limit,
		IsAscending: isAscending,
	}
}

// GetHeadersResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetHeadersResponseMessage struct {
	baseMessage
	Headers []string

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetHeadersResponseMessage) Command() MessageCommand {
	return CmdGetHeadersResponseMessage
}

// NewGetHeadersResponseMessage returns a instance of the message
func NewGetHeadersResponseMessage(headers []string) *GetHeadersResponseMessage {
	return &GetHeadersResponseMessage{
		Headers: headers,
	}
}
