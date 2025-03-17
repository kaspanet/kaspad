package appmessage

// GetPruningWindowRootsRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetPruningWindowRootsRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetPruningWindowRootsRequestMessage) Command() MessageCommand {
	return CmdGetPruningWindowRootsRequestMessage
}

type PruningWindowRoots struct {
	PPRoots []string
	PPIndex uint64
}

// GetPruningWindowRootsResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetPruningWindowRootsResponseMessage struct {
	baseMessage
	Roots []*PruningWindowRoots
	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetPruningWindowRootsResponseMessage) Command() MessageCommand {
	return CmdGetPruningWindowRootsResponseMessage
}
