package appmessage

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

// MsgBlockWithTrustedData represents a kaspa BlockWithTrustedData message
type MsgBlockWithTrustedData struct {
	baseMessage

	Block        *MsgBlock
	DAAScore     uint64
	DAAWindow    []*TrustedDataDataDAABlock
	GHOSTDAGData []*BlockGHOSTDAGDataHashPair
}

// Command returns the protocol command string for the message
func (msg *MsgBlockWithTrustedData) Command() MessageCommand {
	return CmdBlockWithTrustedData
}

// NewMsgBlockWithTrustedData returns a new MsgBlockWithTrustedData.
func NewMsgBlockWithTrustedData() *MsgBlockWithTrustedData {
	return &MsgBlockWithTrustedData{}
}

// TrustedDataDataDAABlock is an appmessage representation of externalapi.TrustedDataDataDAABlock
type TrustedDataDataDAABlock struct {
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
