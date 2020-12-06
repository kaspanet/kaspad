package ghostdagmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockGHOSTDAGData struct {
	blueScore          uint64
	selectedParent     *externalapi.DomainHash
	mergeSetBlues      []*externalapi.DomainHash
	mergeSetReds       []*externalapi.DomainHash
	bluesAnticoneSizes map[externalapi.DomainHash]model.KType
}

// NewBlockGHOSTDAGData creates a new instance of model.BlockGHOSTDAGData
func NewBlockGHOSTDAGData(
	blueScore uint64,
	selectedParent *externalapi.DomainHash,
	mergeSetBlues []*externalapi.DomainHash,
	mergeSetReds []*externalapi.DomainHash,
	bluesAnticoneSizes map[externalapi.DomainHash]model.KType) model.BlockGHOSTDAGData {

	return &blockGHOSTDAGData{
		blueScore:          blueScore,
		selectedParent:     selectedParent,
		mergeSetBlues:      mergeSetBlues,
		mergeSetReds:       mergeSetReds,
		bluesAnticoneSizes: bluesAnticoneSizes,
	}
}

func (bgd *blockGHOSTDAGData) BlueScore() uint64 {
	return bgd.blueScore
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
