package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

type blockGHOSTDAGData struct {
	blueScore          uint64
	blueWork           *big.Int
	selectedParent     *externalapi.DomainHash
	mergeSetBlues      []*externalapi.DomainHash
	mergeSetReds       []*externalapi.DomainHash
	bluesAnticoneSizes map[externalapi.DomainHash]model.KType
}

// NewBlockGHOSTDAGData creates a new instance of model.BlockGHOSTDAGData
func NewBlockGHOSTDAGData(
	blueScore uint64,
	blueWork *big.Int,
	selectedParent *externalapi.DomainHash,
	mergeSetBlues []*externalapi.DomainHash,
	mergeSetReds []*externalapi.DomainHash,
	bluesAnticoneSizes map[externalapi.DomainHash]model.KType) model.BlockGHOSTDAGData {

	return &blockGHOSTDAGData{
		blueScore:          blueScore,
		blueWork:           blueWork,
		selectedParent:     selectedParent,
		mergeSetBlues:      mergeSetBlues,
		mergeSetReds:       mergeSetReds,
		bluesAnticoneSizes: bluesAnticoneSizes,
	}
}

func (bgd *blockGHOSTDAGData) BlueScore() uint64 {
	return bgd.blueScore
}

func (bgd *blockGHOSTDAGData) BlueWork() *big.Int {
	return bgd.blueWork
}

func (bgd *blockGHOSTDAGData) SelectedParent() *externalapi.DomainHash {
	return bgd.selectedParent
}

func (bgd *blockGHOSTDAGData) MergeSetBlues() []*externalapi.DomainHash {
	return bgd.mergeSetBlues
}

func (bgd *blockGHOSTDAGData) MergeSetReds() []*externalapi.DomainHash {
	return bgd.mergeSetReds
}

func (bgd *blockGHOSTDAGData) BluesAnticoneSizes() map[externalapi.DomainHash]model.KType {
	return bgd.bluesAnticoneSizes
}

func (bgd *blockGHOSTDAGData) MergeSet() []*externalapi.DomainHash {
	mergeSet := make([]*externalapi.DomainHash, len(bgd.mergeSetBlues)+len(bgd.mergeSetReds))
	copy(mergeSet, bgd.mergeSetBlues)
	if len(bgd.mergeSetReds) > 0 {
		copy(mergeSet[len(bgd.mergeSetBlues):], bgd.mergeSetReds)
	}

	return mergeSet
}
