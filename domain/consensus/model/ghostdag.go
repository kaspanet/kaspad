package model

import (
	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// BlockGHOSTDAGData represents GHOSTDAG data for some block
type BlockGHOSTDAGData interface {
	BlueScore() uint64
	BlueWork() *big.Int
	SelectedParent() *externalapi.DomainHash
	MergeSetBlues() []*externalapi.DomainHash
	MergeSetReds() []*externalapi.DomainHash
	BluesAnticoneSizes() map[externalapi.DomainHash]KType
	Equal(other BlockGHOSTDAGData) bool
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
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
