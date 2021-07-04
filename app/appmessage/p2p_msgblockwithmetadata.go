package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// MsgBlockWithMetaData represents a kaspa BlockWithMetaData message
type MsgBlockWithMetaData struct {
	baseMessage

	*externalapi.BlockWithMetaData
}

// Command returns the protocol command string for the message
func (msg *MsgBlockWithMetaData) Command() MessageCommand {
	return CmdBlockWithMetaData
}

// NewMsgBlockWithMetaData returns a new MsgBlockWithMetaData.
func NewMsgBlockWithMetaData(blockWithMetaData *externalapi.BlockWithMetaData) *MsgBlockWithMetaData {
	return &MsgBlockWithMetaData{
		BlockWithMetaData: blockWithMetaData,
	}
}
