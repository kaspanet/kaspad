package appmessage

// BanRequestMessage is an appmessage corresponding to
// its respective RPC message
type BanRequestMessage struct {
	baseMessage

	IP string
}

// Command returns the protocol command string for the message
func (msg *BanRequestMessage) Command() MessageCommand {
	return CmdBanRequestMessage
}

// NewBanRequestMessage returns an instance of the message
func NewBanRequestMessage(ip string) *BanRequestMessage {
	return &BanRequestMessage{
		IP: ip,
	}
}

// BanResponseMessage is an appmessage corresponding to
// its respective RPC message
type BanResponseMessage struct {
	baseMessage

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *BanResponseMessage) Command() MessageCommand {
	return CmdBanResponseMessage
}

// NewBanResponseMessage returns a instance of the message
func NewBanResponseMessage() *BanResponseMessage {
	return &BanResponseMessage{}
}
