package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

// MsgBlockWithMetaData represents a kaspa BlockWithMetaData message
type MsgBlockWithMetaData struct {
	baseMessage

	Block        *MsgBlock
	DAAScore     uint64
	GHOSTDAGData *GHOSTDAGData
}

// Command returns the protocol command string for the message
func (msg *MsgBlockWithMetaData) Command() MessageCommand {
	return CmdBlockWithMetaData
}

// NewMsgBlockWithMetaData returns a new MsgBlockWithMetaData.
func NewMsgBlockWithMetaData(block *MsgBlock, daaScore uint64, ghostdagData *GHOSTDAGData) *MsgBlockWithMetaData {
	return &MsgBlockWithMetaData{
		Block:        block,
		DAAScore:     daaScore,
		GHOSTDAGData: ghostdagData,
	}
}

type GHOSTDAGData struct {
	BlueScore          uint64
	BlueWork           *big.Int
	SelectedParent     *externalapi.DomainHash
	MergeSetBlues      []*externalapi.DomainHash
	MergeSetReds       []*externalapi.DomainHash
	BluesAnticoneSizes []*BluesAnticoneSizes
}

type BluesAnticoneSizes struct {
	BlueHash     *externalapi.DomainHash
	AnticoneSize externalapi.KType
}
