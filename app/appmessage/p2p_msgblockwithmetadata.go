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
	DAAWindow    []*BlockWithMetaDataDAABlock
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

// Command returns the protocol command string for the message
func (msg *MsgBlockWithMetaData) Command() MessageCommand {
	return CmdBlockWithMetaData
}

// NewMsgBlockWithMetaData returns a new MsgBlockWithMetaData.
func NewMsgBlockWithMetaData() *MsgBlockWithMetaData {
	return &MsgBlockWithMetaData{}
}

// BlockWithMetaDataDAABlock is an appmessage representation of externalapi.BlockWithMetaDataDAABlock
type BlockWithMetaDataDAABlock struct {
	Header       *MsgBlockHeader
	GHOSTDAGData *BlockGHOSTDAGData
}

// BlockGHOSTDAGData is an appmessage representation of externalapi.BlockGHOSTDAGData
type BlockGHOSTDAGData struct {
	BlueScore          uint64
	BlueWork           *big.Int
	SelectedParent     *externalapi.DomainHash
	MergeSetBlues      []*externalapi.DomainHash
	MergeSetReds       []*externalapi.DomainHash
	BluesAnticoneSizes []*BluesAnticoneSizes
}

// BluesAnticoneSizes is an appmessage representation of the BluesAnticoneSizes part of GHOSTDAG data.
type BluesAnticoneSizes struct {
	BlueHash     *externalapi.DomainHash
	AnticoneSize externalapi.KType
}

// BlockGHOSTDAGDataHashPair is an appmessage representation of externalapi.BlockGHOSTDAGDataHashPair
type BlockGHOSTDAGDataHashPair struct {
	Hash         *externalapi.DomainHash
	GHOSTDAGData *BlockGHOSTDAGData
}
