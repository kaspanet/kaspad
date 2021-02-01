package appmessage

import "net"

// UnbanRequestMessage is an appmessage corresponding to
// its respective RPC message
type UnbanRequestMessage struct {
	baseMessage

	IP net.IP
}

// Command returns the protocol command string for the message
func (msg *UnbanRequestMessage) Command() MessageCommand {
	return CmdUnbanRequestMessage
}

// NewUnbanRequestMessage returns an instance of the message
func NewUnbanRequestMessage(ip net.IP) *UnbanRequestMessage {
	return &UnbanRequestMessage{
		IP: ip,
	}
}

// UnbanResponseMessage is an appmessage corresponding to
// its respective RPC message
type UnbanResponseMessage struct {
	baseMessage

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *UnbanResponseMessage) Command() MessageCommand {
	return CmdUnbanResponseMessage
}

// NewUnbanResponseMessage returns a instance of the message
func NewUnbanResponseMessage() *UnbanResponseMessage {
	return &UnbanResponseMessage{}
}
