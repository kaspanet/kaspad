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

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ = &BlockGHOSTDAGData{0, &externalapi.DomainHash{}, []*externalapi.DomainHash{},
	[]*externalapi.DomainHash{}, map[externalapi.DomainHash]KType{}}

// Equal returns whether bgd equals to other
func (bgd *BlockGHOSTDAGData) Equal(other *BlockGHOSTDAGData) bool {
	if bgd == nil || other == nil {
		return bgd == other
	}

	if bgd.BlueScore != other.BlueScore {
		return false
	}

	if !bgd.SelectedParent.Equal(other.SelectedParent) {
		return false
	}

	if !externalapi.HashesEqual(bgd.MergeSetBlues, other.MergeSetBlues) {
		return false
	}

	if !externalapi.HashesEqual(bgd.MergeSetReds, other.MergeSetReds) {
		return false
	}

	if len(bgd.BluesAnticoneSizes) != len(other.BluesAnticoneSizes) {
		return false
	}

	for hash, size := range bgd.BluesAnticoneSizes {
		otherSize, exists := other.BluesAnticoneSizes[hash]
		if !exists {
			return false
		}

		if size != otherSize {
			return false
		}
	}

	return true
}

// KType defines the size of GHOSTDAG consensus algorithm K parameter.
type KType byte
