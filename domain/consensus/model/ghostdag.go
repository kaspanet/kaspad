package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData struct {
	BlueScore          uint64
	SelectedParent     *externalapi.DomainHash
	MergeSetBlues      []*externalapi.DomainHash
	MergeSetReds       []*externalapi.DomainHash
	BluesAnticoneSizes map[externalapi.DomainHash]KType
}

// Clone returns a clone of BlockGHOSTDAGData
func (bgd *BlockGHOSTDAGData) Clone() *BlockGHOSTDAGData {
	if bgd == nil {
		return nil
	}

	bluesAnticoneSizesClone := make(map[externalapi.DomainHash]KType, len(bgd.BluesAnticoneSizes))
	for hash, size := range bgd.BluesAnticoneSizes {
		bluesAnticoneSizesClone[hash] = size
	}

	return &BlockGHOSTDAGData{
		BlueScore:          bgd.BlueScore,
		SelectedParent:     bgd.SelectedParent.Clone(),
		MergeSetBlues:      externalapi.CloneHashes(bgd.MergeSetBlues),
		MergeSetReds:       externalapi.CloneHashes(bgd.MergeSetReds),
		BluesAnticoneSizes: bluesAnticoneSizesClone,
	}
}

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType byte
