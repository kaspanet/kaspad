package appmessage

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// GetBlockCountRequestMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockCountRequestMessage struct {
	baseMessage
}

// Command returns the protocol command string for the message
func (msg *GetBlockCountRequestMessage) Command() MessageCommand {
	return CmdGetBlockCountRequestMessage
}

// NewGetBlockCountRequestMessage returns a instance of the message
func NewGetBlockCountRequestMessage() *GetBlockCountRequestMessage {
	return &GetBlockCountRequestMessage{}
}

// GetBlockCountResponseMessage is an appmessage corresponding to
// its respective RPC message
type GetBlockCountResponseMessage struct {
	baseMessage
	BlockCount  uint64
	HeaderCount uint64

	Error *RPCError
}

// Command returns the protocol command string for the message
func (msg *GetBlockCountResponseMessage) Command() MessageCommand {
	return CmdGetBlockCountResponseMessage
}

// NewGetBlockCountResponseMessage returns a instance of the message
func NewGetBlockCountResponseMessage(syncInfo *externalapi.SyncInfo) *GetBlockCountResponseMessage {
	return &GetBlockCountResponseMessage{
		BlockCount:  syncInfo.BlockCount,
		HeaderCount: syncInfo.HeaderCount,
	}
}
