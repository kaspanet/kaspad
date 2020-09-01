package appmessage

// ConnectToPeerRequestMessage is an appmessage corresponding to
// its respective RPC message
type ConnectToPeerRequestMessage struct {
	baseMessage
	Address     string
	IsPermanent bool
}

// Command returns the protocol command string for the message
func (msg *ConnectToPeerRequestMessage) Command() MessageCommand {
	return CmdConnectToPeerRequestMessage
}

// ConnectToPeerRequestMessage returns a instance of the message
func NewConnectToPeerRequestMessage(address string, isPermanent bool) *ConnectToPeerRequestMessage {
	return &ConnectToPeerRequestMessage{
		Address:     address,
		IsPermanent: isPermanent,
	}
}

// ConnectToPeerResponseMessage is an appmessage corresponding to
// its respective RPC message
type ConnectToPeerResponseMessage struct {
	baseMessage
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *ConnectToPeerResponseMessage) Command() MessageCommand {
	return CmdConnectToPeerResponseMessage
}

// ConnectToPeerResponseMessage returns a instance of the message
func NewConnectToPeerResponseMessage() *ConnectToPeerResponseMessage {
	return &ConnectToPeerResponseMessage{}
}
